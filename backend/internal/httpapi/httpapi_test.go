package httpapi

import (
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"title":"Book","unknown":true}`))
	rec := httptest.NewRecorder()
	var dst projectInput
	if decode(rec, req, &dst) { t.Fatal("expected invalid request to be rejected") }
	if rec.Code != 400 { t.Fatalf("expected status 400, got %d", rec.Code) }
}

func TestEncryptDecryptTokenRoundTrip(t *testing.T) {
	t.Setenv("TOKEN_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901")))
	encrypted, err := encryptToken("refresh-token-secret")
	if err != nil { t.Fatalf("encrypt token: %v", err) }
	if encrypted == "refresh-token-secret" { t.Fatal("token must not be stored in plaintext") }
	plain, err := decryptToken(encrypted)
	if err != nil { t.Fatalf("decrypt token: %v", err) }
	if plain != "refresh-token-secret" { t.Fatalf("unexpected decrypted value %q", plain) }
}

func TestEncryptTokenRequiresKey(t *testing.T) {
	t.Setenv("TOKEN_ENCRYPTION_KEY", "")
	if _, err := encryptToken("secret"); err == nil { t.Fatal("expected missing encryption key to fail") }
}

func TestFlattenJSONPreservesDocumentOrder(t *testing.T) {
	doc := map[string]any{"type":"doc","content":[]any{
		map[string]any{"type":"paragraph","content":[]any{map[string]any{"type":"text","text":"First"}}},
		map[string]any{"type":"paragraph","content":[]any{map[string]any{"type":"text","text":"Second"}}},
	}}
	if got := flattenJSON(doc); got != "First Second" { t.Fatalf("expected document order, got %q", got) }
}

func TestXMLEscape(t *testing.T) {
	if got := xmlEscape(`A & <B> "C"`); got != "A &amp; &lt;B&gt; &quot;C&quot;" { t.Fatalf("unexpected XML escaping: %q", got) }
}
