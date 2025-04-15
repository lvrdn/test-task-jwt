package handler

import "net/http"

func (h *handler) SetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth", h.issue)
	mux.HandleFunc("GET /api/refresh", h.refresh)
}
