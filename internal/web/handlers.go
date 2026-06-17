package web

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"http-anatomy/internal/inspector"
)

// maxFieldLen caps stored field lengths to limit memory abuse.
const maxFieldLen = 200

// clip trims whitespace and enforces maxFieldLen.
func clip(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > maxFieldLen {
		s = s[:maxFieldLen]
	}
	return s
}

// handleIndex renders the full two-pane page. A fresh visitor gets a session
// cookie here. The inspector starts empty (this is a full load, not an HTMX
// request), so no exchange is recorded.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	sid := sessionID(w, r)
	view := PageView{
		Todos:     s.store.Todos(sid),
		Inspector: InspectorView{OOB: false, History: s.store.History(sid)},
	}
	out, err := renderToString("page", view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, out)
}

// writeWithInspector emits the primary fragment plus the inspector as a
// single response: the inspector carries hx-swap-oob so one request updates
// two regions. The primary string is captured into the exchange first, so
// the inspector never renders its own OOB block.
func (s *Server) writeWithInspector(w http.ResponseWriter, r *http.Request, sid, primary string, status int) {
	ex := inspector.BuildExchange(r, primary, status, w.Header())
	s.store.AppendHistory(sid, ex)
	insp, err := renderToString("inspector", InspectorView{
		OOB:     true,
		Current: &ex,
		History: s.store.History(sid),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, primary+insp)
}

// respondError sends an error fragment as the primary swap, still updating
// the inspector so the failed exchange is visible too.
func (s *Server) respondError(w http.ResponseWriter, r *http.Request, sid string, code int, msg string) {
	primary, err := renderToString("error", ErrorView{Code: code, Message: msg})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, code)
}

func (s *Server) handleTodosFragment(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	primary, err := renderToString("todos", s.store.Todos(sid))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, http.StatusOK)
}

func (s *Server) handleUsersFragment(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	primary, err := renderToString("users", s.store.Users(sid))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, http.StatusOK)
}

func (s *Server) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	title := clip(r.FormValue("title"))
	if title == "" {
		s.respondError(w, r, sid, http.StatusUnprocessableEntity, "title is required")
		return
	}
	t, ok := s.store.AddTodoUnique(sid, title)
	if !ok {
		s.respondError(w, r, sid, http.StatusConflict, "a todo with that title already exists")
		return
	}
	primary, err := renderToString("todo-item", t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, http.StatusCreated)
}

// handlePutTodo fully replaces a todo (PUT, idempotent): both title and done
// come from the form. Contrast with PATCH, which updates a single field.
func (s *Server) handlePutTodo(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.respondError(w, r, sid, http.StatusNotFound, "invalid todo id")
		return
	}
	title := clip(r.FormValue("title"))
	if title == "" {
		s.respondError(w, r, sid, http.StatusUnprocessableEntity, "title is required")
		return
	}
	done := r.FormValue("done") == "true" || r.FormValue("done") == "on"
	t, found := s.store.ReplaceTodo(sid, id, title, done)
	if !found {
		s.respondError(w, r, sid, http.StatusNotFound, "todo not found")
		return
	}
	primary, err := renderToString("todo-item", t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, http.StatusOK)
}

func (s *Server) handlePatchTodo(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.respondError(w, r, sid, http.StatusNotFound, "invalid todo id")
		return
	}
	_ = r.ParseForm()
	// A title field (edit form) means rename; its absence (toggle button)
	// means flip Done.
	if _, isEdit := r.PostForm["title"]; isEdit {
		title := clip(r.PostForm.Get("title"))
		if title == "" {
			s.respondError(w, r, sid, http.StatusUnprocessableEntity, "title is required")
			return
		}
		t, found := s.store.UpdateTodo(sid, id, title)
		if !found {
			s.respondError(w, r, sid, http.StatusNotFound, "todo not found")
			return
		}
		primary, err := renderToString("todo-item", t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.writeWithInspector(w, r, sid, primary, http.StatusOK)
		return
	}
	t, found := s.store.ToggleTodo(sid, id)
	if !found {
		s.respondError(w, r, sid, http.StatusNotFound, "todo not found")
		return
	}
	primary, err := renderToString("todo-item", t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, http.StatusOK)
}

func (s *Server) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || !s.store.DeleteTodo(sid, id) {
		s.respondError(w, r, sid, http.StatusNotFound, "todo not found")
		return
	}
	// Empty primary fragment: the row's outerHTML is replaced with nothing.
	s.writeWithInspector(w, r, sid, "", http.StatusOK)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	name := clip(r.FormValue("name"))
	if name == "" {
		s.respondError(w, r, sid, http.StatusUnprocessableEntity, "name is required")
		return
	}
	email := clip(r.FormValue("email"))
	primary, err := renderToString("user-item", s.store.AddUser(sid, name, email))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, http.StatusCreated)
}

func (s *Server) handlePatchUser(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.respondError(w, r, sid, http.StatusNotFound, "invalid user id")
		return
	}
	name := clip(r.FormValue("name"))
	if name == "" {
		s.respondError(w, r, sid, http.StatusUnprocessableEntity, "name is required")
		return
	}
	email := clip(r.FormValue("email"))
	u, found := s.store.UpdateUser(sid, id, name, email)
	if !found {
		s.respondError(w, r, sid, http.StatusNotFound, "user not found")
		return
	}
	primary, err := renderToString("user-item", u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeWithInspector(w, r, sid, primary, http.StatusOK)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || !s.store.DeleteUser(sid, id) {
		s.respondError(w, r, sid, http.StatusNotFound, "user not found")
		return
	}
	s.writeWithInspector(w, r, sid, "", http.StatusOK)
}
