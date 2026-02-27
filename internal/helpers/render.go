package helpers

import (
	"fmt"
	"log"
	"net/http"
	"url/internal/templates"

	"github.com/CloudyKit/jet/v6"
	"github.com/gin-gonic/gin"
)

// views loads all .jet templates from ./views
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./views"),
	jet.InDevelopmentMode(),
)

//RenderPage renders a jet template and writes it to the response
//c - ResponseWriter
//templateName - filename
//data - TemplateData

func RenderPage(c *gin.Context, templateName string, data *templates.TemplateData) {
	//Build templates data
	var td templates.TemplateData
	if data != nil {
		td = *data
	}

	//Load the template file from ./views
	t, err := views.GetTemplate(fmt.Sprintf("%s.jet", templateName))
	if err != nil {
		log.Println("error loading template:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "template not found",
		})
		return
	}

	//jet.VarMap passes variables directly into the template
	vars := make(jet.VarMap)

	//Execute renders the template into the response writer
	if err = t.Execute(c.Writer, vars, td); err != nil {
		log.Println("error executing template:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"errro": "could not render template",
		})
		return
	}

}
