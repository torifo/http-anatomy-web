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
	if !strings.Contains(body, "POST /api/todos") {
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
	if !strings.Contains(body, "PUT /api/todos/1") {
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
	if !strings.Contains(body, "DELETE /api/todos/1") {
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
