package httpapi

import (
	"net/http"
	"os"
	"strings"
)

type integrationStatus struct {
	ID string `json:"id"`
	Configured bool `json:"configured"`
	Operational bool `json:"operational"`
	Detail string `json:"detail"`
}

func (a *API) integrations(w http.ResponseWriter,r *http.Request) {
	if _,ok:=user(r);!ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	ctx:=r.Context()
	dbOK:=a.DB.Ping(ctx)==nil
	nvidia:=strings.TrimSpace(os.Getenv("NVIDIA_API_KEY"))!=""
	drive:=strings.TrimSpace(os.Getenv("GOOGLE_DRIVE_CLIENT_ID"))!=""&&strings.TrimSpace(os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET"))!=""&&strings.TrimSpace(os.Getenv("GOOGLE_DRIVE_REDIRECT_URL"))!=""
	var driveConnected bool
	if err:=a.DB.QueryRow(ctx,`SELECT EXISTS(SELECT 1 FROM google_drive_connections WHERE owner_id=$1)`,mustUserID(r)).Scan(&driveConnected);err!=nil{internal(w,err);return}
	storage:=strings.TrimSpace(os.Getenv("ASSET_STORAGE_DIR"))!=""
	firebase:=strings.TrimSpace(os.Getenv("FIREBASE_PROJECT_ID"))!=""
	items:=[]integrationStatus{
		{ID:"neon-postgresql",Configured:true,Operational:dbOK,Detail:choose(dbOK,"Database ping succeeded.","Database ping failed.")},
		{ID:"firebase-auth",Configured:firebase,Operational:firebase,Detail:choose(firebase,"Firebase project ID is configured; token verification occurs on protected API requests.","FIREBASE_PROJECT_ID is not configured.")},
		{ID:"nvidia-nim",Configured:nvidia,Operational:false,Detail:choose(nvidia,"API key is configured; provider health is verified when a generation request runs.","NVIDIA_API_KEY is not configured.")},
		{ID:"asset-storage",Configured:storage,Operational:storage,Detail:choose(storage,"Storage directory is configured; actual disk operations are checked per upload.","ASSET_STORAGE_DIR is not configured.")},
		{ID:"google-drive",Configured:drive,Operational:driveConnected,Detail:choose(driveConnected,"OAuth token is stored. A real sync request validates the token and Drive access.",choose(drive,"OAuth credentials are configured but this account is not connected.","Google Drive OAuth credentials are not configured."))},
	}
	jsonOK(w,200,map[string]any{"items":items})
}
func choose(ok bool,yes,no string)string{if ok{return yes};return no}
func mustUserID(r *http.Request) string { id,_:=user(r); return id }
