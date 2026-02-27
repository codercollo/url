package templates

// TemplateData holds all data passed to every template
type TemplateData struct {
	CSRFToken string
	Flash     string
	Data      map[string]interface{} // for any extra data per page
	Error     string                 // for displaying errors in the UI
}
