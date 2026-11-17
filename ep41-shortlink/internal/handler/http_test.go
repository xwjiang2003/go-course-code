package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"shortlink/internal/cache"
	"shortlink/internal/service"
	"shortlink/internal/store"
)

func newTestHandler() *Handler {
	return New(service.New(store.NewMemory(), cache.New()))
}

func TestShorten(t *testing.T) {
	h := newTestHandler()
	body := `{"url":"http://example.com"}`
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Shorten(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp shortenResp
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Code == "" {
		t.Fatal("code should not be empty")
	}
}

func TestRedirect(t *testing.T) {
	h := newTestHandler()
	code, _ := h.svc.Shorten("http://golang.org")

	req := httptest.NewRequest("GET", "/"+code, nil)
	rec := httptest.NewRecorder()
	h.Redirect(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "http://golang.org" {
		t.Fatalf("Location = %q", loc)
	}
}
