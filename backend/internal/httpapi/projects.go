package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"book-studio/backend/internal/auth"
)

type API struct { DB *pgxpool.Pool }

type Project struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Author string `json:"author"`
	Subtitle string `json:"subtitle"`
	Genre string `json:"genre"`
	Language string `json:"language"`
	Format string `json:"format"`
	Description string `json:"description"`
	Status string `json:"status"`
	Content json.RawMessage `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type projectInput struct {
	Title string `json:"title"`
	Author string `json:"author"`
	Subtitle string `json:"subtitle"`
	Genre string `json:"genre"`
	Language string `json:"language"`
	Format string `json:"format"`
	Description string `json:"description"`
	Status string `json:"status"`
	Content json.RawMessage `json:"content"`
}
type chapterInput struct {
	Title string `json:"title"`
	Position int `json:"position"`
	Content json.RawMessage `json:"content"`
}
type Chapter struct {
	ID string `json:"id"`
	ProjectID string `json:"projectId"`
	Title string `json:"title"`
	Position int `json:"position"`
	Content json.RawMessage `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/projects", a.listProjects)
	mux.HandleFunc("POST /api/v1/projects", a.createProject)
	mux.HandleFunc("GET /api/v1/projects/{id}", a.getProject)
	mux.HandleFunc("PATCH /api/v1/projects/{id}", a.updateProject)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", a.deleteProject)
	mux.HandleFunc("GET /api/v1/projects/{id}/chapters", a.listChapters)
	mux.HandleFunc("POST /api/v1/projects/{id}/chapters", a.createChapter)
	mux.HandleFunc("PATCH /api/v1/projects/{id}/chapters/{chapterID}", a.updateChapter)
	mux.HandleFunc("DELETE /api/v1/projects/{id}/chapters/{chapterID}", a.deleteChapter)
	mux.HandleFunc("GET /api/v1/me/preferences", a.getPreferences)
	mux.HandleFunc("PUT /api/v1/me/preferences", a.putPreferences)
	mux.HandleFunc("POST /api/v1/projects/{id}/ai", a.generateAI)
	mux.HandleFunc("GET /api/v1/reviews", a.listReviews)
	mux.HandleFunc("GET /api/v1/dashboard", a.dashboard)
	mux.HandleFunc("GET /api/v1/projects/{id}/export/epub", a.exportEPUB)
	mux.HandleFunc("GET /api/v1/assets", a.listAssets)
	mux.HandleFunc("POST /api/v1/assets", a.uploadAsset)
	mux.HandleFunc("GET /api/v1/assets/{id}/download", a.downloadAsset)
	mux.HandleFunc("DELETE /api/v1/assets/{id}", a.deleteAsset)
}

func user(r *http.Request) (string, bool) { return auth.UserID(r.Context()) }

func (a *API) listProjects(w http.ResponseWriter, r *http.Request) {
	uid, ok := user(r); if !ok { fail(w, 401, "UNAUTHENTICATED", "Sign-in required"); return }
	rows, err := a.DB.Query(r.Context(), `SELECT id::text,title,author,subtitle,genre,language,format,description,status,content,created_at,updated_at FROM projects WHERE owner_id=$1 ORDER BY updated_at DESC`, uid)
	if err != nil { internal(w, err); return }; defer rows.Close()
	items := []Project{}
	for rows.Next() { var p Project; if err:=rows.Scan(&p.ID,&p.Title,&p.Author,&p.Subtitle,&p.Genre,&p.Language,&p.Format,&p.Description,&p.Status,&p.Content,&p.CreatedAt,&p.UpdatedAt); err!=nil { internal(w,err); return }; items=append(items,p) }
	if err:=rows.Err(); err!=nil { internal(w,err); return }; jsonOK(w,200,map[string]any{"items":items,"total":len(items)})
}
func (a *API) createProject(w http.ResponseWriter, r *http.Request) {
	uid,ok:=user(r); if !ok { fail(w,401,"UNAUTHENTICATED","Sign-in required"); return }
	var in projectInput; if !decode(w,r,&in) { return }; in.Title=strings.TrimSpace(in.Title); in.Author=strings.TrimSpace(in.Author)
	if in.Title=="" || in.Author=="" { fail(w,400,"VALIDATION_ERROR","Title and author are required"); return }
	if in.Language=="" { in.Language="pt-BR" }; if in.Format=="" { in.Format="A5" }; if in.Status=="" { in.Status="draft" }; if len(in.Content)==0 { in.Content=json.RawMessage(`{"type":"doc","content":[{"type":"paragraph"}]}`) }
	var p Project
	err:=a.DB.Exec(r.Context(),`INSERT INTO users(id) VALUES($1) ON CONFLICT(id) DO NOTHING`,uid)
	if err!=nil && !errors.Is(err,pgx.ErrNoRows) { internal(w,err); return }
	err=a.DB.QueryRow(r.Context(),`INSERT INTO projects(owner_id,title,author,subtitle,genre,language,format,description,status,content) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb) RETURNING id::text,title,author,subtitle,genre,language,format,description,status,content,created_at,updated_at`,uid,in.Title,in.Author,in.Subtitle,in.Genre,in.Language,in.Format,in.Description,in.Status,string(in.Content)).Scan(&p.ID,&p.Title,&p.Author,&p.Subtitle,&p.Genre,&p.Language,&p.Format,&p.Description,&p.Status,&p.Content,&p.CreatedAt,&p.UpdatedAt)
	if err!=nil { internal(w,err); return }; jsonOK(w,201,p)
}
func (a *API) getProject(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r); if !ok { fail(w,401,"UNAUTHENTICATED","Sign-in required"); return }
	p,err:=a.project(r,uid,r.PathValue("id")); if errors.Is(err,pgx.ErrNoRows) { fail(w,404,"NOT_FOUND","Project not found"); return }; if err!=nil { internal(w,err); return }; jsonOK(w,200,p)
}
func (a *API) project(r *http.Request,uid,id string)(Project,error) {
	if _,err:=uuid.Parse(id); err!=nil { return Project{},pgx.ErrNoRows }
	var p Project
	err:=a.DB.QueryRow(r.Context(),`SELECT id::text,title,author,subtitle,genre,language,format,description,status,content,created_at,updated_at FROM projects WHERE id=$1 AND owner_id=$2`,id,uid).Scan(&p.ID,&p.Title,&p.Author,&p.Subtitle,&p.Genre,&p.Language,&p.Format,&p.Description,&p.Status,&p.Content,&p.CreatedAt,&p.UpdatedAt)
	return p,err
}
func (a *API) updateProject(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r); if !ok { fail(w,401,"UNAUTHENTICATED","Sign-in required"); return }; id:=r.PathValue("id")
	var in projectInput; if !decode(w,r,&in) { return }; if strings.TrimSpace(in.Title)=="" || strings.TrimSpace(in.Author)=="" { fail(w,400,"VALIDATION_ERROR","Title and author are required"); return }; if len(in.Content)==0 { in.Content=json.RawMessage(`{"type":"doc","content":[{"type":"paragraph"}]}`) }; if in.Status=="" { in.Status="draft" }
	var p Project
	err:=a.DB.QueryRow(r.Context(),`UPDATE projects SET title=$3,author=$4,subtitle=$5,genre=$6,language=$7,format=$8,description=$9,status=$10,content=$11::jsonb,updated_at=now() WHERE id=$1 AND owner_id=$2 RETURNING id::text,title,author,subtitle,genre,language,format,description,status,content,created_at,updated_at`,id,uid,strings.TrimSpace(in.Title),strings.TrimSpace(in.Author),in.Subtitle,in.Genre,in.Language,in.Format,in.Description,in.Status,string(in.Content)).Scan(&p.ID,&p.Title,&p.Author,&p.Subtitle,&p.Genre,&p.Language,&p.Format,&p.Description,&p.Status,&p.Content,&p.CreatedAt,&p.UpdatedAt)
	if errors.Is(err,pgx.ErrNoRows) { fail(w,404,"NOT_FOUND","Project not found"); return }; if err!=nil { internal(w,err); return }; jsonOK(w,200,p)
}
func (a *API) deleteProject(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r); if !ok { fail(w,401,"UNAUTHENTICATED","Sign-in required"); return }
	tag,err:=a.DB.Exec(r.Context(),`DELETE FROM projects WHERE id=$1 AND owner_id=$2`,r.PathValue("id"),uid); if err!=nil { internal(w,err); return }; if tag.RowsAffected()==0 { fail(w,404,"NOT_FOUND","Project not found"); return }; w.WriteHeader(http.StatusNoContent)
}
func (a *API) listChapters(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r); if !ok { fail(w,401,"UNAUTHENTICATED","Sign-in required"); return }; id:=r.PathValue("id")
	if _,err:=a.project(r,uid,id); err!=nil { if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","Project not found")}else{internal(w,err)}; return }
	rows,err:=a.DB.Query(r.Context(),`SELECT id::text,project_id::text,title,position,content,created_at,updated_at FROM chapters WHERE project_id=$1 AND owner_id=$2 ORDER BY position,id`,id,uid); if err!=nil { internal(w,err); return }; defer rows.Close(); items:=[]Chapter{}
	for rows.Next(){var c Chapter;if err:=rows.Scan(&c.ID,&c.ProjectID,&c.Title,&c.Position,&c.Content,&c.CreatedAt,&c.UpdatedAt);err!=nil{internal(w,err);return};items=append(items,c)}
	if err:=rows.Err();err!=nil{internal(w,err);return};jsonOK(w,200,map[string]any{"items":items,"total":len(items)})
}
func (a *API) createChapter(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return};id:=r.PathValue("id")
	if _,err:=a.project(r,uid,id);err!=nil{if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","Project not found")}else{internal(w,err)};return}
	var in chapterInput;if !decode(w,r,&in){return};in.Title=strings.TrimSpace(in.Title);if in.Title==""{fail(w,400,"VALIDATION_ERROR","Chapter title is required");return};if len(in.Content)==0{in.Content=json.RawMessage(`{"type":"doc","content":[{"type":"paragraph"}]}`)}
	var c Chapter;err:=a.DB.QueryRow(r.Context(),`INSERT INTO chapters(project_id,owner_id,title,position,content) VALUES($1,$2,$3,(SELECT COALESCE(MAX(position)+1,0) FROM chapters WHERE project_id=$1),$5::jsonb) RETURNING id::text,project_id::text,title,position,content,created_at,updated_at`,id,uid,in.Title,in.Position,string(in.Content)).Scan(&c.ID,&c.ProjectID,&c.Title,&c.Position,&c.Content,&c.CreatedAt,&c.UpdatedAt)
	if err!=nil{internal(w,err);return};jsonOK(w,201,c)
}
func (a *API) updateChapter(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return};var in chapterInput;if !decode(w,r,&in){return};if strings.TrimSpace(in.Title)==""||len(in.Content)==0{fail(w,400,"VALIDATION_ERROR","Chapter title and content are required");return}
	var c Chapter;err:=a.DB.QueryRow(r.Context(),`UPDATE chapters SET title=$4,position=$5,content=$6::jsonb,updated_at=now() WHERE id=$1 AND project_id=$2 AND owner_id=$3 RETURNING id::text,project_id::text,title,position,content,created_at,updated_at`,r.PathValue("chapterID"),r.PathValue("id"),uid,strings.TrimSpace(in.Title),in.Position,string(in.Content)).Scan(&c.ID,&c.ProjectID,&c.Title,&c.Position,&c.Content,&c.CreatedAt,&c.UpdatedAt)
	if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","Chapter not found");return};if err!=nil{internal(w,err);return};jsonOK(w,200,c)
}
func (a *API) deleteChapter(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return};tag,err:=a.DB.Exec(r.Context(),`DELETE FROM chapters WHERE id=$1 AND project_id=$2 AND owner_id=$3`,r.PathValue("chapterID"),r.PathValue("id"),uid);if err!=nil{internal(w,err);return};if tag.RowsAffected()==0{fail(w,404,"NOT_FOUND","Chapter not found");return};w.WriteHeader(http.StatusNoContent)
}
func (a *API) getPreferences(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return};var raw json.RawMessage;err:=a.DB.QueryRow(r.Context(),`SELECT preferences FROM user_preferences WHERE owner_id=$1`,uid).Scan(&raw);if errors.Is(err,pgx.ErrNoRows){jsonOK(w,200,map[string]any{"preferences":map[string]any{}});return};if err!=nil{internal(w,err);return};jsonOK(w,200,map[string]any{"preferences":raw})
}
func (a *API) putPreferences(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return};var body struct{Preferences json.RawMessage `json:"preferences"`};if !decode(w,r,&body){return};if len(body.Preferences)==0||!json.Valid(body.Preferences){fail(w,400,"VALIDATION_ERROR","A valid preferences object is required");return}
	_,err:=a.DB.Exec(r.Context(),`INSERT INTO user_preferences(owner_id,preferences) VALUES($1,$2::jsonb) ON CONFLICT(owner_id) DO UPDATE SET preferences=EXCLUDED.preferences,updated_at=now()`,uid,string(body.Preferences));if err!=nil{internal(w,err);return};jsonOK(w,200,map[string]any{"preferences":body.Preferences})
}
func decode(w http.ResponseWriter,r *http.Request,dst any) bool {
	r.Body=http.MaxBytesReader(w,r.Body,2<<20);dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
	if err:=dec.Decode(dst);err!=nil{fail(w,400,"INVALID_JSON","Request body is invalid");return false};return true
}
func jsonOK(w http.ResponseWriter,status int,payload any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(payload)}
func fail(w http.ResponseWriter,status int,code,message string){jsonOK(w,status,map[string]any{"error":map[string]string{"code":code,"message":message}})}
func internal(w http.ResponseWriter,err error){_ = err;fail(w,500,"INTERNAL_ERROR","The request could not be completed")}
