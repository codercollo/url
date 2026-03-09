package handler

import (
	"net/http"
	"url/internal/helpers"
	"url/internal/templates"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to shorten url"})
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
