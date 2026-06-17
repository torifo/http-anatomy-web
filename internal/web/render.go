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

// TodosView feeds the todos pane: the (filtered) list plus the current
// search/filter state and counts for the summary.
type TodosView struct {
	Todos  []model.Todo
	Query  string
	Filter string // "all" | "active" | "done"
	Total  int    // all todos in the session (pre-filter)
	Done   int    // completed todos in the session (pre-filter)
}

// PageView feeds the full two-pane page on GET /.
type PageView struct {
	Theme     string // "dark" | "light"
	Todos     TodosView
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
