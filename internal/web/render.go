package web

import (
	"embed"
	"html/template"
	"strings"

	"http-anatomy/internal/model"
)

//go:embed templates/*.html
var templatesFS embed.FS

var tmpl = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

// InspectorView feeds the inspector pane. OOB marks the fragment for an
// out-of-band swap (true on CRUD/tab responses, false on full page load).
type InspectorView struct {
	OOB     bool
	Current *model.Exchange
	History []model.Exchange
}

// PageView feeds the full two-pane page on GET /.
type PageView struct {
	Todos     []model.Todo
	Inspector InspectorView
}

// ErrorView feeds the error fragment.
type ErrorView struct {
	Code    int
	Message string
}

// renderToString executes a named template into a string.
func renderToString(name string, data any) (string, error) {
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, name, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
