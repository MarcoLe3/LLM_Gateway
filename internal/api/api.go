package api

import (
	"net/http"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /job", h.submit())
	return mux
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	
}