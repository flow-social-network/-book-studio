package httpapi

import (
  "context"
  "net/http"
  "time"
  "github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
 ID string `json:"id"`
 Title string `json:"title"`
 Message string `json:"message"`
 Href string `json:"href"`
 ReadAt *time.Time `json:"readAt"`
 CreatedAt time.Time `json:"createdAt"`
}

func createNotification(ctx context.Context, db *pgxpool.Pool, ownerID, title, message, href string) error {
 _, err := db.Exec(ctx, `INSERT INTO notifications(owner_id,title,message,href) VALUES($1,$2,$3,$4)`, ownerID, title, message, href)
 return err
}
func (a *API) listNotifications(w http.ResponseWriter, r *http.Request) {
 uid, ok := user(r)
 if !ok { fail(w, 401, "UNAUTHENTICATED", "Sign-in required"); return }
 rows, err := a.DB.Query(r.Context(), `SELECT id::text,title,message,href,read_at,created_at FROM notifications WHERE owner_id=$1 ORDER BY created_at DESC LIMIT 50`, uid)
 if err != nil { internal(w, err); return }
 defer rows.Close()
 items := []Notification{}
 var unread int64
 for rows.Next() {
  var n Notification
  if err := rows.Scan(&n.ID,&n.Title,&n.Message,&n.Href,&n.ReadAt,&n.CreatedAt); err != nil { internal(w,err); return }
  if n.ReadAt == nil { unread++ }
  items = append(items,n)
 }
 if err := rows.Err(); err != nil { internal(w,err); return }
 jsonOK(w,200,map[string]any{"items":items,"unreadCount":unread})
}
func (a *API) markNotificationRead(w http.ResponseWriter, r *http.Request) {
 uid, ok := user(r)
 if !ok { fail(w,401,"UNAUTHENTICATED","Sign-in required"); return }
 tag,err:=a.DB.Exec(r.Context(),`UPDATE notifications SET read_at=COALESCE(read_at,now()) WHERE id=$1 AND owner_id=$2`,r.PathValue("id"),uid)
 if err != nil { internal(w,err); return }
 if tag.RowsAffected()==0 { fail(w,404,"NOT_FOUND","Notification not found"); return }
 w.WriteHeader(http.StatusNoContent)
}
