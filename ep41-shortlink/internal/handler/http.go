package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"shortlink/internal/service"
)

type Handler struct {
	svc *service.Shortener
}

func New(svc *service.Shortener) *Handler {
	return &Handler{svc: svc}
}

type shortenReq struct {
	URL string `json:"url"`
}

type shortenResp struct {
	Code string `json:"code"`
	URL  string `json:"url"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	code, err := h.svc.Shorten(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortenResp{Code: code, URL: req.URL})
}

// Redirect 处理 GET /{code}，命中短码后 302 重定向。
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	url, err := h.svc.Resolve(code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}
