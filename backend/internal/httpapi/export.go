package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (a *API) exportEPUB(w http.ResponseWriter,r *http.Request) {
	uid,ok:=user(r);if !ok{fail(w,401,"UNAUTHENTICATED","Sign-in required");return}
	p,err:=a.project(r,uid,r.PathValue("id"));if errors.Is(err,pgx.ErrNoRows){fail(w,404,"NOT_FOUND","Project not found");return};if err!=nil{internal(w,err);return}
	chapters,err:=a.DB.Query(r.Context(),`SELECT title,content FROM chapters WHERE project_id=$1 AND owner_id=$2 ORDER BY position,id`,p.ID,uid);if err!=nil{internal(w,err);return}
	type chapter struct{Title string;Content json.RawMessage}; list:=[]chapter{}
	for chapters.Next(){var c chapter;if err:=chapters.Scan(&c.Title,&c.Content);err!=nil{chapters.Close();internal(w,err);return};list=append(list,c)}
	if err:=chapters.Err();err!=nil{chapters.Close();internal(w,err);return};chapters.Close()
	var b bytes.Buffer;zw:=zip.NewWriter(&b)
	write:=func(name,content string)error{f,err:=zw.Create(name);if err!=nil{return err};_,err=f.Write([]byte(content));return err}
	mimeHeader:=&zip.FileHeader{Name:"mimetype",Method:zip.Store};mimeWriter,err:=zw.CreateHeader(mimeHeader);if err!=nil{internal(w,err);return};if _,err=mimeWriter.Write([]byte("application/epub+zip"));err!=nil{internal(w,err);return}
	if err:=write("META-INF/container.xml",`<?xml version="1.0" encoding="UTF-8"?><container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles></container>`);err!=nil{internal(w,err);return}
	items:=[]string{};spine:=[]string{};manifest:=""
	makeChapter:=func(title string,content json.RawMessage)string{var node any;_ = json.Unmarshal(content,&node);text:=flattenJSON(node);return "<?xml version=\"1.0\" encoding=\"UTF-8\"?><html xmlns=\"http://www.w3.org/1999/xhtml\"><head><title>"+xmlEscape(title)+"</title></head><body><h1>"+xmlEscape(title)+"</h1><p>"+strings.ReplaceAll(xmlEscape(text),"\n","</p><p>")+"</p></body></html>"}
	items=append(items,"<item id=\"title\" href=\"title.xhtml\" media-type=\"application/xhtml+xml\"/>");spine=append(spine,"<itemref idref=\"title\"/>")
	if err:=write("OEBPS/title.xhtml",makeChapter(p.Title,json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","text":""}]}`)));err!=nil{internal(w,err);return}
	for i,c:=range list{id:=fmt.Sprintf("chapter-%d",i+1);file:=id+".xhtml";items=append(items,fmt.Sprintf("<item id=\"%s\" href=\"%s\" media-type=\"application/xhtml+xml\"/>",id,file));spine=append(spine,fmt.Sprintf("<itemref idref=\"%s\"/>",id));if err:=write("OEBPS/"+file,makeChapter(c.Title,c.Content));err!=nil{internal(w,err);return}}
	for _,item:=range items{manifest+=item};spineXML:=strings.Join(spine,"")
	opf:=`<?xml version="1.0" encoding="UTF-8"?><package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id"><metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:identifier id="book-id">`+xmlEscape(p.ID)+`</dc:identifier><dc:title>`+xmlEscape(p.Title)+`</dc:title><dc:creator>`+xmlEscape(p.Author)+`</dc:creator><dc:language>`+xmlEscape(p.Language)+`</dc:language></metadata><manifest>`+manifest+`</manifest><spine>`+spineXML+`</spine></package>`
	if err:=write("OEBPS/content.opf",opf);err!=nil{internal(w,err);return};if err:=zw.Close();err!=nil{internal(w,err);return}
	w.Header().Set("Content-Type","application/epub+zip");w.Header().Set("Content-Disposition",`attachment; filename="book.epub"`);w.WriteHeader(http.StatusOK);_,_=w.Write(b.Bytes())
}
func flattenJSON(v any)string{switch x:=v.(type){case map[string]any:if t,ok:=x["text"].(string);ok{return t};if children,ok:=x["content"];ok{return flattenJSON(children)};return "";case []any:var out []string;for _,v:=range x{if s:=flattenJSON(v);s!=""{out=append(out,s)}};return strings.Join(out," ");default:return ""}}
func xmlEscape(s string)string{r:=strings.NewReplacer("&","&amp;","<","&lt;",">","&gt;","\"","&quot;","'","&apos;");return r.Replace(s)}
