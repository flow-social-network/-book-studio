package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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
type nimMessage struct { Role string `json:"role"`; Content string `json:"content"` }
type nimResponse struct { Choices []struct { Message nimMessage `json:"message"` } `json:"choices"`; Error *struct{Message string `json:"message"`} `json:"error"` }

func (a *API) generateAI(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	instruction:="";var in aiRequest;if r.Body!=nil&&r.ContentLength!=0{if !decode(w,r,&in){return};instruction=strings.TrimSpace(in.Instruction)}
	if instruction==""{fail(w,400,"VALIDATION_ERROR","Instruction is required");return}
	p,err:=a.project(r,uid,r.PathValue("id"));if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","Project not found");return};if err!=nil{internal(w,err);return}
	content,_:=json.Marshal(p.Content)
	system:="Você é um assistente editorial profissional. Ajude o autor sem afirmar fatos não presentes no material. Preserve a autoria, não invente referências e devolva apenas o resultado solicitado. Contexto do livro: título="+p.Title+"; autor="+p.Author+"; gênero="+p.Genre+"; descrição="+p.Description+"; manuscrito JSON="+string(content)
	if in.Mode=="review"{system="Você é um revisor editorial. Analise apenas o material fornecido e produza uma revisão estruturada em português com seções: pontos fortes, problemas identificados, sugestões concretas e dúvidas para o autor. Não altere o manuscrito automaticamente. Contexto do livro: "+p.Title+"; manuscrito JSON="+string(content)}
	result,err:=callNVIDIA(r.Context(),system,instruction);if err!=nil{fail(w,502,"AI_PROVIDER_ERROR",err.Error());return}
	if in.Mode=="review" {
		resultJSON,_:=json.Marshal(map[string]any{"result":result,"instruction":instruction})
		_,err=a.DB.Exec(r.Context(),`INSERT INTO editorial_reviews(owner_id,project_id,status,provider,result,completed_at) VALUES($1,$2,'completed','nvidia-nim',$3::jsonb,now())`,uid,p.ID,string(resultJSON))
		if err!=nil{internal(w,err);return}
	}
	jsonOK(w,200,map[string]any{"projectId":p.ID,"mode":in.Mode,"provider":"nvidia-nim","model":envOr("NVIDIA_MODEL","meta/llama-3.1-70b-instruct"),"result":result,"generatedAt":time.Now().UTC()})
}
func callNVIDIA(ctx context.Context,system,user string)(string,error){
	key:=strings.TrimSpace(os.Getenv("NVIDIA_API_KEY"));if key==""{return "",fmt.Errorf("NVIDIA_API_KEY is not configured")}
	model:=envOr("NVIDIA_MODEL","meta/llama-3.1-70b-instruct");base:=strings.TrimRight(envOr("NVIDIA_BASE_URL","https://integrate.api.nvidia.com/v1"),"/")
	payload:=nimRequest{Model:model,Messages:[]nimMessage{{Role:"system",Content:system},{Role:"user",Content:user}},Temperature:0.4,MaxTokens:3000}
	data,err:=json.Marshal(payload);if err!=nil{return "",err}
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,base+"/chat/completions",bytes.NewReader(data));if err!=nil{return "",err};req.Header.Set("Authorization","Bearer "+key);req.Header.Set("Content-Type","application/json")
	client:=&http.Client{Timeout:90*time.Second};resp,err:=client.Do(req);if err!=nil{return "",fmt.Errorf("NVIDIA request failed: %w",err)};defer resp.Body.Close()
	body,err:=io.ReadAll(io.LimitReader(resp.Body,4<<20));if err!=nil{return "",err};if resp.StatusCode<200||resp.StatusCode>=300{return "",fmt.Errorf("NVIDIA provider returned HTTP %d",resp.StatusCode)}
	var parsed nimResponse;if err:=json.Unmarshal(body,&parsed);err!=nil{return "",fmt.Errorf("NVIDIA response was not valid JSON")};if parsed.Error!=nil{return "",fmt.Errorf("NVIDIA provider returned an error")};if len(parsed.Choices)==0||strings.TrimSpace(parsed.Choices[0].Message.Content)==""{return "",fmt.Errorf("NVIDIA provider returned no content")};return parsed.Choices[0].Message.Content,nil
}
func envOr(key,fallback string)string{if v:=strings.TrimSpace(os.Getenv(key));v!=""{return v};return fallback}
