package web

import (
	"strings"

	"http-anatomy/internal/model"
)

// buildTodosView filters todos by query (case-insensitive title match) and
// filter ("all"|"active"|"done"), and computes pre-filter counts for the
// summary. An unknown filter is treated as "all".
func buildTodosView(all []model.Todo, query, filter string) TodosView {
	switch filter {
	case "active", "done":
	default:
		filter = "all"
	}
	q := strings.ToLower(strings.TrimSpace(query))

	var done int
	filtered := make([]model.Todo, 0, len(all))
	for _, t := range all {
		if t.Done {
			done++
		}
		if filter == "active" && t.Done {
			continue
		}
		if filter == "done" && !t.Done {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(t.Title), q) {
			continue
		}
		filtered = append(filtered, t)
	}
	return TodosView{
		Todos:  filtered,
		Query:  query,
		Filter: filter,
		Total:  len(all),
		Done:   done,
	}
}
