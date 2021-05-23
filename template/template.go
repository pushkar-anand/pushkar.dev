package template

import (
	"html/template"
	"net/http"

	"go.uber.org/zap"
)

type Renderer struct {
	templates *template.Template
	logger    *zap.Logger
}

//NewRenderer creates a new template Renderer
func NewRenderer(templates *template.Template, logger *zap.Logger) *Renderer {
	return &Renderer{
		templates: templates,
		logger:    logger,
	}
}

//Render the given HTML template
func (r *Renderer) Render(w http.ResponseWriter, templateName string, templateData interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := r.templates.ExecuteTemplate(w, templateName, templateData)
	if err != nil {
		r.logger.With(zap.Error(err), zap.String("template", templateName)).Error("error rendering template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
