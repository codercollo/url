package handler

import (
	"errors"
	"net/http"
	"url/internal/helpers"
	"url/internal/templates"
	"url/pkg/sanitize"

	"github.com/gin-gonic/gin"
)

type ShortenRequest struct {
	URL        string `json:"url" form:"url" binding:"required"`
	CustomCode string `json:"custom_code" form:"custom_code"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func (h *Handler) Shorten(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	url, err := h.shortener.Shorten(c.Request.Context(), req.URL, req.CustomCode, "", nil)
	if err != nil {
		msg := "Something went wrong, please try again"
		switch {
		case errors.Is(err, sanitize.ErrEmptyURL):
			msg = "Please enter a URL"
		case errors.Is(err, sanitize.ErrInvalidURL):
			msg = "That doesn't look like a valid URL"
		case errors.Is(err, sanitize.ErrUnsafeScheme):
			msg = "only http and https URLS are allowed"
		case errors.Is(err, sanitize.ErrMissingHost):
			msg = "URL must include a domain name"
		case errors.Is(err, sanitize.ErrPrivateHost),
			errors.Is(err, sanitize.ErrLoopbackHost):
			msg = "URLS pointing to private or local addresses are not allowed "
		}

		//Form clients get HTML error, API clients get JSON
		if c.ContentType() == "application/x-www-form-urlencoded" {
			helpers.RenderPage(c, "home", &templates.TemplateData{
				Error: msg,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": msg,
		})
		return

	}

	shortURL := "http://localhost:8080/" + url.ShortCode

	// Form submissions always send application/x-www-form-urlencoded
	// Use that as the reliable signal to render HTML instead of JSON
	if c.ContentType() == "application/x-www-form-urlencoded" {
		helpers.RenderPage(c, "home", &templates.TemplateData{
			Data: map[string]interface{}{
				"short_url": shortURL,
			},
		})
		return
	}

	// API clients get JSON
	c.JSON(http.StatusOK, ShortenResponse{
		ShortCode: url.ShortCode,
		ShortURL:  shortURL,
	})
}
