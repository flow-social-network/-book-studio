package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"title":"Book","unknown":true}`))
	rec := httptest.NewRecorder()
	var dst projectInput
	if decode(rec, req, &dst) {
		t.Fatal("expected invalid request to be rejected")
	}
	if rec.Code != 400 {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
