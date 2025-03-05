package view

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

var (
	templates map[string]*template.Template
	initialized bool
)

// Initialize loads all templates
func Initialize() error {
	if initialized {
		return nil
	}

	templates = make(map[string]*template.Template)

	// Define template functions
	funcMap := template.FuncMap{
		// Add any custom functions here if needed
	}

	// Load all page templates
	pages, err := filepath.Glob(filepath.Join("templates", "*.html"))
	if err != nil {
		return err
	}

	// Parse each page template
	for _, page := range pages {
		name := filepath.Base(page)
		// Skip layout template, it will be included with each page
		if name == "layout.html" {
			continue
		}

		// Remove the .html extension
		name = name[:len(name)-5]

		// Create new template with the functions
		tmpl := template.New(name).Funcs(funcMap)

		// First parse the layout template
		layoutTmpl, err := template.ParseFiles(filepath.Join("templates", "layout.html"))
		if err != nil {
			return err
		}

		// Clone the layout template
		tmpl, err = layoutTmpl.Clone()
		if err != nil {
			return err
		}

		// Parse the page content into the layout
		tmpl, err = tmpl.ParseFiles(page)
		if err != nil {
			return err
		}

		templates[name] = tmpl
	}

	initialized = true
	return nil
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// Initialize templates if not already done
	if !initialized {
		if err := Initialize(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize templates")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Get the template
	tmpl, ok := templates[name]
	if !ok {
		log.Error().Str("template", name).Msg("Template not found")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the template to a buffer first to catch any rendering errors
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		log.Error().Err(err).Str("template", name).Msg("Failed to render template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set content type and write the rendered template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}