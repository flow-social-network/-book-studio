package httpapi

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/jackc/pgx/v5"
)

func driveOAuthConfig() (*oauth2.Config,error) {
	id:=strings.TrimSpace(os.Getenv("GOOGLE_DRIVE_CLIENT_ID"))
	secret:=strings.TrimSpace(os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET"))
	redirect:=strings.TrimSpace(os.Getenv("GOOGLE_DRIVE_REDIRECT_URL"))
	if id==""||secret==""||redirect=="" { return nil,fmt.Errorf("Google Drive OAuth credentials and redirect URL are not configured") }
	return &oauth2.Config{ClientID:id,ClientSecret:secret,RedirectURL:redirect,Scopes:[]string{"https://www.googleapis.com/auth/drive.file"},Endpoint:google.Endpoint},nil
}

func (a *API) driveStatus(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	var connected bool;var connectedAt *time.Time
	err:=a.DB.QueryRow(r.Context(),`SELECT true,connected_at FROM google_drive_connections WHERE owner_id=$1`,uid).Scan(&connected,&connectedAt)
	if errors.Is(err,pgx.ErrNoRows){connected=false}else if err!=nil{internal(w,err);return}
	jsonOK(w,200,map[string]any{"connected":connected,"connectedAt":connectedAt})
}

func (a *API) startDriveOAuth(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	cfg,err:=driveOAuthConfig();if err!=nil{fail(w,503,"DRIVE_OAUTH_NOT_CONFIGURED",err.Error());return}
	raw:=make([]byte,32);if _,err:=rand.Read(raw);err!=nil{internal(w,err);return}
	state:=base64.RawURLEncoding.EncodeToString(raw);sum:=sha256.Sum256([]byte(state));hash:=hex.EncodeToString(sum[:])
	if _,err:=a.DB.Exec(r.Context(),`INSERT INTO users(id) VALUES($1) ON CONFLICT(id) DO NOTHING`,uid);err!=nil{internal(w,err);return}
	if _,err:=a.DB.Exec(r.Context(),`DELETE FROM oauth_states WHERE expires_at < now()`);err!=nil{internal(w,err);return}
	if _,err:=a.DB.Exec(r.Context(),`INSERT INTO oauth_states(state_hash,owner_id,expires_at) VALUES($1,$2,now()+interval '10 minutes')`,hash,uid);err!=nil{internal(w,err);return}
	jsonOK(w,200,map[string]string{"authorizationUrl":cfg.AuthCodeURL(state,oauth2.AccessTypeOffline,oauth2.ApprovalForce)})
}

func (a *API) GoogleDriveCallback(w http.ResponseWriter,r *http.Request) {
	if providerErr:=r.URL.Query().Get("error");providerErr!=""{http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=consent_denied",http.StatusSeeOther);return}
	state:=r.URL.Query().Get("state");code:=r.URL.Query().Get("code")
	if state==""||code==""{http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=invalid_callback",http.StatusSeeOther);return}
	sum:=sha256.Sum256([]byte(state));hash:=hex.EncodeToString(sum[:])
	var uid string
	err:=a.DB.QueryRow(r.Context(),`DELETE FROM oauth_states WHERE state_hash=$1 AND expires_at>now() RETURNING owner_id`,hash).Scan(&uid)
	if errors.Is(err,pgx.ErrNoRows){http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=invalid_state",http.StatusSeeOther);return};if err!=nil{http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=server_error",http.StatusSeeOther);return}
	cfg,err:=driveOAuthConfig();if err!=nil{http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=not_configured",http.StatusSeeOther);return}
	token,err:=cfg.Exchange(r.Context(),code);if err!=nil||token.RefreshToken==""{http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=token_exchange_failed",http.StatusSeeOther);return}
	encrypted,err:=encryptToken(token.RefreshToken);if err!=nil{http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=encryption_not_configured",http.StatusSeeOther);return}
	_,err=a.DB.Exec(r.Context(),`INSERT INTO google_drive_connections(owner_id,refresh_token_encrypted,scope) VALUES($1,$2,$3) ON CONFLICT(owner_id) DO UPDATE SET refresh_token_encrypted=EXCLUDED.refresh_token_encrypted,scope=EXCLUDED.scope,updated_at=now()`,uid,encrypted,strings.Join(cfg.Scopes," "))
	if err!=nil{http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?error=storage_failed",http.StatusSeeOther);return}
	http.Redirect(w,r,frontendURL()+"/app/integracoes/google-drive?connected=1",http.StatusSeeOther)
}

func (a *API) syncDriveProject(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	var in struct{ProjectID string `json:"projectId"`};if !decode(w,r,&in){return};if strings.TrimSpace(in.ProjectID)==""{fail(w,400,"VALIDATION_ERROR","projectId is required");return}
	cfg,err:=driveOAuthConfig();if err!=nil{fail(w,503,"DRIVE_OAUTH_NOT_CONFIGURED",err.Error());return}
	var encrypted string;err=a.DB.QueryRow(r.Context(),`SELECT refresh_token_encrypted FROM google_drive_connections WHERE owner_id=$1`,uid).Scan(&encrypted)
	if errors.Is(err,pgx.ErrNoRows){fail(w,409,"DRIVE_NOT_CONNECTED","Connect Google Drive before syncing.");return};if err!=nil{internal(w,err);return}
	refresh,err:=decryptToken(encrypted);if err!=nil{fail(w,500,"TOKEN_DECRYPTION_FAILED","Could not decrypt the stored Google Drive token.");return}
	p,err:=a.project(r,uid,in.ProjectID);if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","Project not found");return};if err!=nil{internal(w,err);return}
	rows,err:=a.DB.Query(r.Context(),`SELECT title,content FROM chapters WHERE project_id=$1 AND owner_id=$2 ORDER BY position,id`,p.ID,uid);if err!=nil{internal(w,err);return}
	var body strings.Builder;body.WriteString("# "+p.Title+"\n\nBy "+p.Author+"\n\n"+p.Description+"\n\n")
	for rows.Next(){var title string;var raw []byte;if err:=rows.Scan(&title,&raw);err!=nil{rows.Close();internal(w,err);return};var node any;_ = json.Unmarshal(raw,&node);body.WriteString("\n## "+title+"\n\n"+flattenJSON(node)+"\n")}
	if err:=rows.Err();err!=nil{rows.Close();internal(w,err);return};rows.Close()
	source:=cfg.TokenSource(r.Context(),&oauth2.Token{RefreshToken:refresh,TokenType:"Bearer"})
	token,err:=source.Token();if err!=nil{fail(w,502,"DRIVE_TOKEN_REFRESH_FAILED","Google Drive authorization expired or was revoked. Reconnect the integration.");return}
	fileID,err:=uploadDriveText(r.Context(),token.AccessToken,p.Title+".txt",body.String());if err!=nil{fail(w,502,"DRIVE_UPLOAD_FAILED",err.Error());return}
	_,err=a.DB.Exec(r.Context(),`INSERT INTO google_drive_sync_files(owner_id,project_id,drive_file_id,drive_name,synced_at) VALUES($1,$2,$3,$4,now()) ON CONFLICT(owner_id,project_id) DO UPDATE SET drive_file_id=EXCLUDED.drive_file_id,drive_name=EXCLUDED.drive_name,synced_at=now()`,uid,p.ID,fileID,p.Title+".txt")
	if err!=nil{internal(w,err);return}
	jsonOK(w,200,map[string]any{"synced":true,"projectId":p.ID,"driveFileId":fileID,"syncedAt":time.Now().UTC()})
}

func uploadDriveText(ctx context.Context,accessToken,name,content string)(string,error){
	randomBoundary:=make([]byte,18);if _,err:=rand.Read(randomBoundary);err!=nil{return "",err};boundary:="bookstudio-"+base64.RawURLEncoding.EncodeToString(randomBoundary)
	var body bytes.Buffer
	body.WriteString("--"+boundary+"\r\nContent-Type: application/json; charset=UTF-8\r\n\r\n")
	meta,_:=json.Marshal(map[string]string{"name":name,"mimeType":"text/plain"})
	body.Write(meta);body.WriteString("\r\n--"+boundary+"\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n");body.WriteString(content);body.WriteString("\r\n--"+boundary+"--")
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,"https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id",bytes.NewReader(body.Bytes()));if err!=nil{return "",err}
	req.Header.Set("Authorization","Bearer "+accessToken);req.Header.Set("Content-Type","multipart/related; boundary="+boundary)
	client:=&http.Client{Timeout:45*time.Second};resp,err:=client.Do(req);if err!=nil{return "",err};defer resp.Body.Close()
	raw,err:=io.ReadAll(io.LimitReader(resp.Body,1<<20));if err!=nil{return "",err};if resp.StatusCode<200||resp.StatusCode>=300{return "",fmt.Errorf("Google Drive returned HTTP %d",resp.StatusCode)}
	var result struct{ID string `json:"id"`};if err:=json.Unmarshal(raw,&result);err!=nil||result.ID==""{return "",fmt.Errorf("Google Drive did not return a file ID")};return result.ID,nil
}

func encryptToken(value string)(string,error){
	raw:=strings.TrimSpace(os.Getenv("TOKEN_ENCRYPTION_KEY"));key,err:=base64.StdEncoding.DecodeString(raw);if err!=nil||len(key)!=32{return "",fmt.Errorf("TOKEN_ENCRYPTION_KEY must be base64-encoded 32 bytes")}
	block,err:=aes.NewCipher(key);if err!=nil{return "",err};gcm,err:=cipher.NewGCM(block);if err!=nil{return "",err};nonce:=make([]byte,gcm.NonceSize());if _,err:=rand.Read(nonce);err!=nil{return "",err};sealed:=gcm.Seal(nonce,nonce,[]byte(value),nil);return base64.StdEncoding.EncodeToString(sealed),nil
}
func decryptToken(value string)(string,error){
	raw:=strings.TrimSpace(os.Getenv("TOKEN_ENCRYPTION_KEY"));key,err:=base64.StdEncoding.DecodeString(raw);if err!=nil||len(key)!=32{return "",fmt.Errorf("TOKEN_ENCRYPTION_KEY must be base64-encoded 32 bytes")}
	sealed,err:=base64.StdEncoding.DecodeString(value);if err!=nil{return "",err};block,err:=aes.NewCipher(key);if err!=nil{return "",err};gcm,err:=cipher.NewGCM(block);if err!=nil{return "",err};if len(sealed)<gcm.NonceSize(){return "",fmt.Errorf("invalid encrypted token")};plain,err:=gcm.Open(nil,sealed[:gcm.NonceSize()],sealed[gcm.NonceSize():],nil);if err!=nil{return "",err};return string(plain),nil
}
func frontendURL()string{v:=strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")),"/");if v==""{return "http://localhost:3000"};return v}
