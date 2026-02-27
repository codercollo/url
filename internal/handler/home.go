package handler

import (
	"url/internal/helpers"
	"url/internal/templates"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Home(c *gin.Context) {
	helpers.RenderPage(c, "home", &templates.TemplateData{})
}
