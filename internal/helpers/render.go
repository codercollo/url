package helpers

import (
	"html/template"
	"log"
	"net/http"
	"url/internal/templates"

	"github.com/gin-gonic/gin"
)

type pageConfig struct {
	layout   string
	partials []string
	page     string
}

var pages = map[string]pageConfig{
	"home": {
		layout:   "views/layouts/base.html",
		partials: []string{"views/partials/toast.html", "views/partials/footer.html"},
		page:     "views/home.html",
	},
	"login": {
		layout:   "views/layouts/auth.html",
		partials: []string{"views/partials/toast.html"},
		page:     "views/login.html",
	},
	"create_account": {
		layout:   "views/layouts/auth.html",
		partials: []string{"views/partials/toast.html"},
		page:     "views/create_account.html",
	},
	"admin_stats": {
		layout:   "views/layouts/admin.html",
		partials: []string{"views/partials/toast.html", "views/partials/nav.html", "views/partials/footer.html"},
		page:     "views/admin_stats.html",
	},
}

var layoutNames = map[string]string{
	"views/layouts/base.html":  "base",
	"views/layouts/auth.html":  "auth",
	"views/layouts/admin.html": "admin",
}

func RenderPage(c *gin.Context, templateName string, data *templates.TemplateData) {
	var td templates.TemplateData
	if data != nil {
		td = *data
	}

	cfg, ok := pages[templateName]
	if !ok {
		log.Printf("no page config for: %s", templateName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "template not found"})
		return
	}

	// Build file list: layout + partials + page
	files := []string{cfg.layout}
	files = append(files, cfg.partials...)
	files = append(files, cfg.page)

	// Fresh parse per request — prevents define namespace collision
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		log.Println("error parsing template:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "template parse error"})
		return
	}

	layoutName := layoutNames[cfg.layout]

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, layoutName, td); err != nil {
		log.Println("error executing template:", err)
	}
}
