package handler

import (
	"errors"
	"net/http"
	apperrors "url/internal/errors"

	"github.com/gin-gonic/gin"
)

// Redirect handles GET /:code
// Resolves short code to original URL and issues a 301 redirect
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "code is required",
		})
	}
	url, err := h.shortener.Resolve(c.Request.Context(), code)
	if err != nil {
		//Map service errors to HTTP responses
		if errors.Is(err, apperrors.ErrURLNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "short code not found",
			})
		}
		if errors.Is(err, apperrors.ErrURLExpired) {
			c.JSON(http.StatusGone, gin.H{
				"error": "this link has expired",
			})

		}
		if errors.Is(err, apperrors.ErrURLInactive) {
			c.JSON(http.StatusGone, gin.H{
				"error": "this link is no longer active",
			})
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to resolve url",
		})
		return
	}

	//301 - permanent redirect
	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}
