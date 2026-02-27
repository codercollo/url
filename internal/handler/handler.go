package handler

import "url/internal/service"

//Handler holds all dependencies for HTTP handlers
//Add more services here as the project grows
type Handler struct {
	shortener *service.ShortenerService
}

//NewHandler creates a new Handler with all required services injected
func NewHandler(shortener *service.ShortenerService) *Handler {
	return &Handler{shortener: shortener}
}
