package web

import (
	"strings"
	"testing"

	"http-anatomy/internal/model"
)

func TestRenderTodoItemEscapesAndIDs(t *testing.T) {
	out, err := renderToString("todo-item", model.Todo{ID: 7, Title: "<script>x</script>", Done: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `id="todo-7"`) {
		t.Fatalf("missing id: %s", out)
	}
	if strings.Contains(out, "<script>x</script>") {
		t.Fatalf("title not escaped: %s", out)
	}
	if !strings.Contains(out, "hx-delete=\"/api/todos/7\"") {
		t.Fatalf("missing hx-delete: %s", out)
	}
}

func TestRenderInspectorOOB(t *testing.T) {
	view := InspectorView{
		OOB:     true,
		Current: &model.Exchange{Method: "DELETE", Path: "/api/todos/7", Proto: "HTTP/1.1", Status: 200, StatusText: "OK", Body: "<li>gone</li>"},
		History: []model.Exchange{{Method: "DELETE", Path: "/api/todos/7", Status: 200}},
	}
	out, err := renderToString("inspector", view)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `id="http-inspector"`) {
		t.Fatalf("missing inspector id: %s", out)
	}
	if !strings.Contains(out, `hx-swap-oob="true"`) {
		t.Fatalf("OOB attribute missing: %s", out)
	}
	// The fragment body must be shown escaped, not as live markup.
	if strings.Contains(out, "<li>gone</li>") {
		t.Fatalf("body should be escaped: %s", out)
	}
	if !strings.Contains(out, "&lt;li&gt;gone&lt;/li&gt;") {
		t.Fatalf("escaped body missing: %s", out)
	}
}

func TestRenderInspectorNoOOBOnPage(t *testing.T) {
	out, err := renderToString("inspector", InspectorView{OOB: false})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "hx-swap-oob") {
		t.Fatalf("page-embedded inspector must not be OOB: %s", out)
	}
}
