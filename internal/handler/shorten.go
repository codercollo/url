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

	//ShouldBind handles both JSON and HTML form submissions
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	//Call service - no custom TTL or user owernership for now
	url, err := h.shortener.Shorten(c.Request.Context(), req.URL, req.CustomCode, "", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to shorten url",
		})
		return
	}

	//Build the full short URL
	shortURL := "http://localhost:8080/" + url.ShortCode

	//If request came from the HTML form, re-render the page
	//with the result
	if c.GetHeader("Accept") == "text/html" || c.ContentType() == "application/x-www-form-urlencoded" {
		helpers.RenderPage(c, "home", &templates.TemplateData{
			Data: map[string]interface{}{
				"short_url": shortURL,
			},
		})
		return
	}

	//Otherwise reurn JSON for API clients
	c.JSON(http.StatusOK, ShortenResponse{
		ShortCode: url.ShortCode,
		ShortURL:  shortURL,
	})
}
