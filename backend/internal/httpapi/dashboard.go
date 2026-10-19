package httpapi

import (
	"net/http"
	"time"
)

func (a *API) dashboard(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	var projects,chapters,assets,reviews int64
	if err:=a.DB.QueryRow(r.Context(),`SELECT count(*) FROM projects WHERE owner_id=$1`,uid).Scan(&projects);err!=nil{internal(w,err);return}
	if err:=a.DB.QueryRow(r.Context(),`SELECT count(*) FROM chapters WHERE owner_id=$1`,uid).Scan(&chapters);err!=nil{internal(w,err);return}
	if err:=a.DB.QueryRow(r.Context(),`SELECT count(*) FROM assets WHERE owner_id=$1`,uid).Scan(&assets);err!=nil{internal(w,err);return}
	if err:=a.DB.QueryRow(r.Context(),`SELECT count(*) FROM editorial_reviews WHERE owner_id=$1`,uid).Scan(&reviews);err!=nil{internal(w,err);return}
	rows,err:=a.DB.Query(r.Context(),`SELECT id::text,title,status,updated_at FROM projects WHERE owner_id=$1 ORDER BY updated_at DESC LIMIT 5`,uid);if err!=nil{internal(w,err);return};defer rows.Close()
	type recentProject struct{ID string `json:"id"`;Title string `json:"title"`;Status string `json:"status"`;UpdatedAt time.Time `json:"updatedAt"`}
	recent:=[]recentProject{};for rows.Next(){var p recentProject;if err:=rows.Scan(&p.ID,&p.Title,&p.Status,&p.UpdatedAt);err!=nil{internal(w,err);return};recent=append(recent,p)};if err:=rows.Err();err!=nil{internal(w,err);return}
	jsonOK(w,200,map[string]any{"metrics":map[string]int64{"projects":projects,"chapters":chapters,"assets":assets,"reviews":reviews},"recentProjects":recent})
}
