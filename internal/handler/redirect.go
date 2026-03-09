package handler

import (
	"errors"
	"net/http"
	apperrors "url/internal/errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	url, err := h.shortener.Resolve(c.Request.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrURLNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "short code not found"})
		case errors.Is(err, apperrors.ErrURLExpired):
			c.JSON(http.StatusGone, gin.H{"error": "this link has expired"})
		case errors.Is(err, apperrors.ErrURLInactive):
			c.JSON(http.StatusGone, gin.H{"error": "this link is no longer active"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve url"})
		}
		return
	}

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}
