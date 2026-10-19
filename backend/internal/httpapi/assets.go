package httpapi

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const maxAssetBytes = 12 << 20

type Asset struct {
	ID string `json:"id"`
	ProjectID *string `json:"projectId"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
	ByteSize int64 `json:"byteSize"`
	CreatedAt time.Time `json:"createdAt"`
}

func (a *API) listAssets(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	rows,err:=a.DB.Query(r.Context(),`SELECT id::text,project_id::text,filename,mime_type,byte_size,created_at FROM assets WHERE owner_id=$1 ORDER BY created_at DESC`,uid);if err!=nil{internal(w,err);return};defer rows.Close()
	items:=[]Asset{};for rows.Next(){var x Asset;var projectID *string;if err:=rows.Scan(&x.ID,&projectID,&x.Filename,&x.MimeType,&x.ByteSize,&x.CreatedAt);err!=nil{internal(w,err);return};x.ProjectID=projectID;items=append(items,x)};if err:=rows.Err();err!=nil{internal(w,err);return};jsonOK(w,200,map[string]any{"items":items,"total":len(items)})
}
func (a *API) uploadAsset(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	storageDir:=strings.TrimSpace(os.Getenv("ASSET_STORAGE_DIR"));if storageDir==""{fail(w,503,"STORAGE_NOT_CONFIGURED","Configure ASSET_STORAGE_DIR on persistent backend storage before uploading files.");return}
	r.Body=http.MaxBytesReader(w,r.Body,maxAssetBytes+1<<20);if err:=r.ParseMultipartForm(maxAssetBytes);err!=nil{fail(w,413,"FILE_TOO_LARGE","The upload exceeds the configured size limit.");return}
	file,header,err:=r.FormFile("file");if err!=nil{fail(w,400,"FILE_REQUIRED","Attach a file in the 'file' field.");return};defer file.Close()
	if header.Size<=0||header.Size>maxAssetBytes{fail(w,413,"FILE_TOO_LARGE","File must be between 1 byte and 12 MB.");return}
	name:=filepath.Base(strings.TrimSpace(header.Filename));if name==""||name=="."||name==string(filepath.Separator){fail(w,400,"INVALID_FILENAME","Filename is invalid.");return}
	ext:=strings.ToLower(filepath.Ext(name));allowed:=map[string]string{".png":"image/png",".jpg":"image/jpeg",".jpeg":"image/jpeg",".webp":"image/webp",".pdf":"application/pdf"}
	expected,ok:=allowed[ext];if !ok{fail(w,415,"UNSUPPORTED_FILE_TYPE","Supported types: PNG, JPEG, WebP and PDF.");return}
	buffer:=make([]byte,512);n,err:=file.Read(buffer);if err!=nil&&err!=io.EOF{internal(w,err);return};detected:=http.DetectContentType(buffer[:n]);if !strings.HasPrefix(detected,expected)&&!(expected=="image/jpeg"&&strings.HasPrefix(detected,"image/jpeg")){fail(w,415,"MIME_MISMATCH","The file content does not match its extension.");return}
	if _,err:=file.Seek(0,io.SeekStart);err!=nil{internal(w,err);return}
	projectID:=strings.TrimSpace(r.FormValue("projectId"));var project any
	if projectID!=""{if _,err:=a.project(r,uid,projectID);err!=nil{if errors.Is(err,pgx.ErrNoRows){fail(w,404,"PROJECT_NOT_FOUND","Project not found.");return};internal(w,err);return};project=projectID}
	key:=uuid.NewString()+ext;folder:=filepath.Clean(storageDir);if err:=os.MkdirAll(folder,0700);err!=nil{internal(w,err);return};dest:=filepath.Join(folder,key)
	out,err:=os.OpenFile(dest,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600);if err!=nil{internal(w,err);return}
	written,copyErr:=io.Copy(out,io.LimitReader(file,maxAssetBytes+1));closeErr:=out.Close();if copyErr!=nil||closeErr!=nil||written>maxAssetBytes{_ = os.Remove(dest);fail(w,500,"STORAGE_WRITE_FAILED","Could not store the uploaded file.");return}
	var x Asset;err=a.DB.QueryRow(r.Context(),`INSERT INTO assets(owner_id,project_id,filename,mime_type,storage_key,byte_size) VALUES($1,$2,$3,$4,$5,$6) RETURNING id::text,project_id::text,filename,mime_type,byte_size,created_at`,uid,project,name,expected,key,written).Scan(&x.ID,&x.ProjectID,&x.Filename,&x.MimeType,&x.ByteSize,&x.CreatedAt)
	if err!=nil{_ = os.Remove(dest);internal(w,err);return};jsonOK(w,201,x)
}
func (a *API) downloadAsset(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return};storageDir:=strings.TrimSpace(os.Getenv("ASSET_STORAGE_DIR"));if storageDir==""{fail(w,503,"STORAGE_NOT_CONFIGURED","Persistent file storage is not configured.");return}
	var name,mimeType,key string;err:=a.DB.QueryRow(r.Context(),`SELECT filename,mime_type,storage_key FROM assets WHERE id=$1 AND owner_id=$2`,r.PathValue("id"),uid).Scan(&name,&mimeType,&key);if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","File not found.");return};if err!=nil{internal(w,err);return}
	if filepath.Base(key)!=key{fail(w,500,"INVALID_STORAGE_KEY","Stored file metadata is invalid.");return};f,err:=os.Open(filepath.Join(storageDir,key));if err!=nil{fail(w,404,"FILE_MISSING","The file is missing from configured storage.");return};defer f.Close()
	w.Header().Set("Content-Type",mime.TypeByExtension(filepath.Ext(name)));w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("Content-Disposition",fmt.Sprintf("attachment; filename=%q",name));http.ServeContent(w,r,name,time.Time{},f)
}
func (a *API) deleteAsset(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return};storageDir:=strings.TrimSpace(os.Getenv("ASSET_STORAGE_DIR"));if storageDir==""{fail(w,503,"STORAGE_NOT_CONFIGURED","Persistent file storage is not configured.");return}
	var key string;err:=a.DB.QueryRow(r.Context(),`DELETE FROM assets WHERE id=$1 AND owner_id=$2 RETURNING storage_key`,r.PathValue("id"),uid).Scan(&key);if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","File not found.");return};if err!=nil{internal(w,err);return};if filepath.Base(key)!=key{fail(w,500,"INVALID_STORAGE_KEY","Stored file metadata is invalid.");return};if err:=os.Remove(filepath.Join(storageDir,key));err!=nil&&!os.IsNotExist(err){fail(w,500,"STORAGE_DELETE_FAILED","File metadata was removed but the stored file could not be deleted.");return};w.WriteHeader(http.StatusNoContent)
}
