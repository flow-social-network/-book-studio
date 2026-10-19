package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type aiRequest struct {
	Instruction string `json:"instruction"`
	Mode string `json:"mode"`
}
type nimRequest struct {
	Model string `json:"model"`
	Messages []nimMessage `json:"messages"`
	Temperature float64 `json:"temperature"`
	MaxTokens int `json:"max_tokens"`
}
type nimMessage struct {
	Role string `json:"role"`
	Content string `json:"content"`
}
type nimResponse struct {
	Choices []struct { Message nimMessage `json:"message"` } `json:"choices"`
	Error *struct { Message string `json:"message"` } `json:"error"`
}

func (a *API) generateAI(w http.ResponseWriter, r *http.Request) {
	uid, ok := user(r)
	if !ok { fail(w, 401, "UNAUTHENTICATED", "Sign-in required"); return }
	var in aiRequest
	if !decode(w, r, &in) { return }
	in.Instruction = strings.TrimSpace(in.Instruction)
	if in.Instruction == "" { fail(w, 400, "VALIDATION_ERROR", "Instruction is required"); return }
	if in.Mode != "write" && in.Mode != "review" { fail(w, 400, "VALIDATION_ERROR", "Mode must be write or review"); return }
	p, err := a.project(r, uid, r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) { fail(w, 404, "NOT_FOUND", "Project not found"); return }
	if err != nil { internal(w, err); return }
	manuscript := flattenJSON(p.Content)
	rows, err := a.DB.Query(r.Context(), `SELECT title,content FROM chapters WHERE project_id=$1 AND owner_id=$2 ORDER BY position,id`, p.ID, uid)
	if err != nil { internal(w, err); return }
	var chapterText strings.Builder
	for rows.Next() { var title string; var raw []byte; if err:=rows.Scan(&title,&raw);err!=nil{rows.Close();internal(w,err);return};var node any;_ = json.Unmarshal(raw,&node);chapterText.WriteString("\\n\\n# "+title+"\\n"+flattenJSON(node)) }
	if err:=rows.Err();err!=nil{rows.Close();internal(w,err);return};rows.Close()
	if chapterText.Len()>0 { manuscript += chapterText.String() }
	if len(manuscript)>100000 { manuscript=manuscript[:100000]+"\\n[Conteúdo truncado para respeitar o limite de contexto]" }
	system := "Você é um assistente editorial profissional. Ajude o autor sem afirmar fatos não presentes no material. Preserve a autoria e não invente referências. Contexto do livro: título=" + p.Title + "; autor=" + p.Author + "; gênero=" + p.Genre + "; descrição=" + p.Description + "; manuscrito="+manuscript
	if in.Mode == "review" {
		system = "Você é um revisor editorial. Analise apenas o material fornecido e produza uma revisão estruturada em português com pontos fortes, problemas identificados, sugestões concretas e dúvidas para o autor. Não altere o manuscrito automaticamente. Contexto do livro: " + p.Title + "; manuscrito=" + manuscript
	}
	result, err := callNVIDIA(r.Context(), system, in.Instruction)
	if err != nil {
		logSafe := "NVIDIA provider request failed"
		_ = logSafe
		fail(w, 502, "AI_PROVIDER_ERROR", err.Error())
		return
	}
	if in.Mode == "review" {
		resultJSON, marshalErr := json.Marshal(map[string]any{"result": result, "instruction": in.Instruction})
		if marshalErr != nil { internal(w, marshalErr); return }
		_, err = a.DB.Exec(r.Context(), `INSERT INTO editorial_reviews(owner_id,project_id,status,provider,result,completed_at) VALUES($1,$2,'completed','nvidia-nim',$3::jsonb,now())`, uid, p.ID, string(resultJSON))
		if err != nil { internal(w, err); return }
		if err:=createNotification(r.Context(),a.DB,uid,"Revisão editorial concluída",p.Title,"/app/editor/revisao");err!=nil{log.Printf("create review notification: %v",err)}
	}
	jsonOK(w, 200, map[string]any{"projectId": p.ID, "mode": in.Mode, "provider": "nvidia-nim", "model": envOr("NVIDIA_MODEL", "meta/llama-3.1-70b-instruct"), "result": result, "generatedAt": time.Now().UTC()})
}

func callNVIDIA(ctx context.Context, system, userPrompt string) (string, error) {
	key := strings.TrimSpace(os.Getenv("NVIDIA_API_KEY"))
	if key == "" { return "", fmt.Errorf("NVIDIA_API_KEY is not configured") }
	model := envOr("NVIDIA_MODEL", "meta/llama-3.1-70b-instruct")
	base := strings.TrimRight(envOr("NVIDIA_BASE_URL", "https://integrate.api.nvidia.com/v1"), "/")
	payload := nimRequest{Model: model, Messages: []nimMessage{{Role: "system", Content: system}, {Role: "user", Content: userPrompt}}, Temperature: 0.4, MaxTokens: 3000}
	data, err := json.Marshal(payload)
	if err != nil { return "", err }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(data))
	if err != nil { return "", err }
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return "", fmt.Errorf("NVIDIA request failed: %w", err) }
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil { return "", err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return "", fmt.Errorf("NVIDIA provider returned HTTP %d", resp.StatusCode) }
	var parsed nimResponse
	if err := json.Unmarshal(body, &parsed); err != nil { return "", fmt.Errorf("NVIDIA response was not valid JSON") }
	if parsed.Error != nil { return "", fmt.Errorf("NVIDIA provider returned an error") }
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" { return "", fmt.Errorf("NVIDIA provider returned no content") }
	return parsed.Choices[0].Message.Content, nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" { return v }
	return fallback
}

func (a *API) listReviews(w http.ResponseWriter, r *http.Request) {
	uid, ok := user(r)
	if !ok { fail(w, 401, "UNAUTHENTICATED", "Sign-in required"); return }
	projectID := r.URL.Query().Get("projectId")
	if projectID != "" {
		if _, err := a.project(r, uid, projectID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) { fail(w, 404, "NOT_FOUND", "Project not found"); return }
			internal(w, err); return
		}
	}
	rows, err := a.DB.Query(r.Context(), `SELECT id::text,project_id::text,status,provider,result,error_code,created_at,completed_at FROM editorial_reviews WHERE owner_id=$1 AND ($2='' OR project_id::text=$2) ORDER BY created_at DESC LIMIT 100`, uid, projectID)
	if err != nil { internal(w, err); return }
	defer rows.Close()
	type review struct {
		ID string `json:"id"`
		ProjectID string `json:"projectId"`
		Status string `json:"status"`
		Provider string `json:"provider"`
		Result json.RawMessage `json:"result"`
		ErrorCode *string `json:"errorCode"`
		CreatedAt time.Time `json:"createdAt"`
		CompletedAt *time.Time `json:"completedAt"`
	}
	items := []review{}
	for rows.Next() {
		var item review
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.Status, &item.Provider, &item.Result, &item.ErrorCode, &item.CreatedAt, &item.CompletedAt); err != nil { internal(w, err); return }
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { internal(w, err); return }
	jsonOK(w, 200, map[string]any{"items": items, "total": len(items)})
}
