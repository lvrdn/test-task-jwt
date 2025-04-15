package handler

import (
	"app/internal/sender"
	"app/internal/storage"
	"net/http"
)

type handler struct {
	AccessKey         string
	AccessExpMinutes  int
	RefreshExpMinutes int
	Storage           storage.Storage
	EmailSender       sender.EmailSender
}

func NewHandler(
	accessTokenKey string,
	accessTokenExpMinutes, refreshTokenExpMinutes int,
	st storage.Storage,
	emailSender sender.EmailSender,
) *handler {
	return &handler{
		AccessKey:         accessTokenKey,
		AccessExpMinutes:  accessTokenExpMinutes,
		RefreshExpMinutes: refreshTokenExpMinutes,
		Storage:           st,
		EmailSender:       emailSender,
	}
}

func (h *handler) issue(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test issue"))
}

func (h *handler) refresh(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test refresh"))
}
