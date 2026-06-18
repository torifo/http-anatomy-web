package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"http-anatomy/internal/store"
)

func newTestServer() http.Handler {
	return NewServer(store.New())
}

// do sends a request and returns the recorder. If cookie is non-empty it is
// sent as the session cookie.
func do(t *testing.T, h http.Handler, method, path, cookie string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	r := httptest.NewRequest(method, path, body)
	if form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	r.Header.Set("HX-Request", "true")
	if cookie != "" {
		r.Header.Set("Cookie", sessionCookie+"="+cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func sessionFrom(w *httptest.ResponseRecorder) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == sessionCookie {
			return c.Value
		}
	}
	return ""
}

func TestIndexSetsCookie(t *testing.T) {
	h := newTestServer()
	w := do(t, h, "GET", "/", "", nil)
	if w.Code != 200 {
		t.Fatalf("want 200 got %d", w.Code)
	}
	if sessionFrom(w) == "" {
		t.Fatal("GET / should issue a session cookie")
	}
	if !strings.Contains(w.Body.String(), "http-inspector") {
		t.Fatal("page should contain inspector")
	}
}

func TestCreateTodoReturnsPrimaryPlusOOB(t *testing.T) {
	h := newTestServer()
	w := do(t, h, "POST", "/api/todos", "sess-1", url.Values{"title": {"buy milk"}})
	if w.Code != 201 {
		t.Fatalf("want 201 got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "buy milk") {
		t.Fatal("primary fragment missing todo title")
	}
	if !strings.Contains(body, `id="todo-1"`) {
		t.Fatal("primary fragment missing todo id")
	}
	if !strings.Contains(body, `hx-swap-oob="true"`) {
		t.Fatal("response missing OOB inspector")
	}
	if !strings.Contains(body, ">POST<") || !strings.Contains(body, "/api/todos") {
		t.Fatal("inspector should show the request line")
	}
}

func TestCreateTodoRequiresTitle(t *testing.T) {
	h := newTestServer()
	w := do(t, h, "POST", "/api/todos", "sess-1", url.Values{"title": {"  "}})
	if w.Code != 422 {
		t.Fatalf("want 422 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "title is required") {
		t.Fatal("missing validation message")
	}
}

func TestToggleVsEditTodo(t *testing.T) {
	h := newTestServer()
	do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"task"}})

	// No title field -> toggle Done.
	w := do(t, h, "PATCH", "/api/todos/1", "s", url.Values{})
	if w.Code != 200 || !strings.Contains(w.Body.String(), "done") {
		t.Fatalf("toggle should mark done: %d %s", w.Code, w.Body.String())
	}
	// Title field -> rename.
	w = do(t, h, "PATCH", "/api/todos/1", "s", url.Values{"title": {"renamed"}})
	if !strings.Contains(w.Body.String(), "renamed") {
		t.Fatal("edit should rename todo")
	}
}

func TestNewVisitorGetsSeedExamples(t *testing.T) {
	h := newTestServer()
	w := do(t, h, "GET", "/", "", nil) // no cookie -> first-time visitor
	if !strings.Contains(w.Body.String(), "HTMX を解剖してみる") {
		t.Fatal("first visit should seed an example todo")
	}
	sid := sessionFrom(w)
	wu := do(t, h, "GET", "/fragments/users", sid, nil)
	if !strings.Contains(wu.Body.String(), "Ada Lovelace") {
		t.Fatal("first visit should seed an example user")
	}
	// Returning visitor (cookie present) is not re-seeded after clearing.
	do(t, h, "DELETE", "/api/todos/1", sid, nil)
	w2 := do(t, h, "GET", "/", sid, nil)
	if strings.Contains(w2.Body.String(), "HTMX を解剖してみる") {
		t.Fatal("returning visitor should not be re-seeded")
	}
}

func TestThemeToggle(t *testing.T) {
	h := newTestServer()
	if w := do(t, h, "GET", "/", "", nil); !strings.Contains(w.Body.String(), "theme-dark") {
		t.Fatal("default theme should be dark")
	}
	w := do(t, h, "GET", "/theme/toggle", "", nil)
	if w.Header().Get("HX-Redirect") != "/" {
		t.Fatalf("toggle should set HX-Redirect: %q", w.Header().Get("HX-Redirect"))
	}
	var theme string
	for _, c := range w.Result().Cookies() {
		if c.Name == "ha_theme" {
			theme = c.Value
		}
	}
	if theme != "light" {
		t.Fatalf("toggle should set ha_theme=light, got %q", theme)
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Cookie", "ha_theme=light")
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !strings.Contains(rw.Body.String(), "theme-light") {
		t.Fatal("light cookie should render theme-light")
	}
}

func TestCreateTodoSetsToastTrigger(t *testing.T) {
	h := newTestServer()
	w := do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"task"}})
	trig := w.Header().Get("HX-Trigger")
	if !strings.Contains(trig, "showToast") || !strings.Contains(trig, "created") {
		t.Fatalf("HX-Trigger should fire showToast(created): %q", trig)
	}
	// The same header is surfaced in the inspector response-headers section.
	if !strings.Contains(w.Body.String(), "HX-Trigger") {
		t.Fatal("inspector should display the HX-Trigger response header")
	}
}

func TestDuplicateTodoReturns409(t *testing.T) {
	h := newTestServer()
	do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"task"}})
	w := do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"TASK"}})
	if w.Code != 409 {
		t.Fatalf("want 409 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "already exists") {
		t.Fatal("missing conflict message")
	}
}

func TestPutTodoFullReplace(t *testing.T) {
	h := newTestServer()
	do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"old"}})
	w := do(t, h, "PUT", "/api/todos/1", "s", url.Values{"title": {"new"}, "done": {"true"}})
	if w.Code != 200 {
		t.Fatalf("want 200 got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "new") || !strings.Contains(body, "done") {
		t.Fatalf("PUT should replace title and set done: %s", body)
	}
	if !strings.Contains(body, ">PUT<") || !strings.Contains(body, "/api/todos/1") {
		t.Fatal("inspector should record the PUT")
	}
	w = do(t, h, "PUT", "/api/todos/999", "s", url.Values{"title": {"x"}})
	if w.Code != 404 {
		t.Fatalf("PUT missing id want 404 got %d", w.Code)
	}
}

func TestDeleteTodoEmptyPrimary(t *testing.T) {
	h := newTestServer()
	do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"task"}})
	w := do(t, h, "DELETE", "/api/todos/1", "s", nil)
	if w.Code != 200 {
		t.Fatalf("want 200 got %d", w.Code)
	}
	body := w.Body.String()
	if strings.Contains(body, `id="todo-1"`) {
		t.Fatal("deleted row should not be in primary fragment")
	}
	if !strings.Contains(body, ">DELETE<") || !strings.Contains(body, "/api/todos/1") {
		t.Fatal("inspector should record the delete")
	}
}

func TestMissingTodoReturns404(t *testing.T) {
	h := newTestServer()
	w := do(t, h, "DELETE", "/api/todos/999", "s", nil)
	if w.Code != 404 {
		t.Fatalf("want 404 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "todo not found") {
		t.Fatal("missing 404 message")
	}
}

func TestSessionStatePersistsAcrossRequests(t *testing.T) {
	h := newTestServer()
	do(t, h, "POST", "/api/todos", "keep", url.Values{"title": {"a"}})
	do(t, h, "POST", "/api/todos", "keep", url.Values{"title": {"b"}})
	w := do(t, h, "GET", "/fragments/todos", "keep", nil)
	body := w.Body.String()
	if !strings.Contains(body, ">a<") && !strings.Contains(body, "value=\"a\"") {
		t.Fatalf("first todo missing after reload: %s", body)
	}
	if !strings.Contains(body, "value=\"b\"") {
		t.Fatalf("second todo missing after reload: %s", body)
	}
}

func TestSearchAndFilterTodos(t *testing.T) {
	h := newTestServer()
	do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"buy milk"}})
	do(t, h, "POST", "/api/todos", "s", url.Values{"title": {"write code"}})
	do(t, h, "PATCH", "/api/todos/1", "s", url.Values{}) // mark "buy milk" done

	// Search narrows by title.
	w := do(t, h, "GET", "/fragments/todos?q=milk", "s", nil)
	body := w.Body.String()
	if !strings.Contains(body, "buy milk") || strings.Contains(body, "write code") {
		t.Fatalf("search q=milk should show only milk: %s", body)
	}
	// Summary counts are pre-filter (2 total, 1 done).
	if !strings.Contains(body, "2 件中 1 完了") {
		t.Fatalf("summary wrong: %s", body)
	}
	// filter=active hides the done one.
	w = do(t, h, "GET", "/fragments/todos?filter=active", "s", nil)
	body = w.Body.String()
	if strings.Contains(body, "buy milk") || !strings.Contains(body, "write code") {
		t.Fatalf("filter=active should hide done todo: %s", body)
	}
}

func TestSessionIsolationAcrossCookies(t *testing.T) {
	h := newTestServer()
	do(t, h, "POST", "/api/todos", "alice", url.Values{"title": {"alice-task"}})
	w := do(t, h, "GET", "/fragments/todos", "bob", nil)
	if strings.Contains(w.Body.String(), "alice-task") {
		t.Fatal("bob must not see alice's todos")
	}
}

func TestUserCRUDFlow(t *testing.T) {
	h := newTestServer()
	w := do(t, h, "POST", "/api/users", "s", url.Values{"name": {"Ada"}, "email": {"ada@x.dev"}})
	if w.Code != 201 || !strings.Contains(w.Body.String(), "Ada") {
		t.Fatalf("create user failed: %d %s", w.Code, w.Body.String())
	}
	w = do(t, h, "PATCH", "/api/users/1", "s", url.Values{"name": {"Ada L."}, "email": {"ada.l@x.dev"}})
	if !strings.Contains(w.Body.String(), "Ada L.") {
		t.Fatal("update user failed")
	}
	w = do(t, h, "DELETE", "/api/users/1", "s", nil)
	if w.Code != 200 {
		t.Fatalf("delete user failed: %d", w.Code)
	}
}
