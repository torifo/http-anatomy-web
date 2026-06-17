// Package store holds per-session, in-memory state for http-anatomy.
// State is volatile: it lives only for the lifetime of the process.
package store

import (
	"strings"
	"sync"

	"http-anatomy/internal/model"
)

// maxHistory caps the inspector log kept per session.
const maxHistory = 10

// Session is one visitor's isolated state.
type Session struct {
	Todos   []model.Todo
	Users   []model.User
	History []model.Exchange // newest first, capped at maxHistory
	todoSeq int
	userSeq int
}

// Store maps session IDs to their state. All access is mutex-guarded;
// Sessions are never handed out, so callers mutate only through Store
// methods and cannot race on a Session's fields.
type Store struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

// New returns an empty Store.
func New() *Store {
	return &Store{sessions: make(map[string]*Session)}
}

// ensure returns the session for id, creating it if absent.
// Callers must hold s.mu.
func (s *Store) ensure(id string) *Session {
	sess := s.sessions[id]
	if sess == nil {
		sess = &Session{}
		s.sessions[id] = sess
	}
	return sess
}

// Has reports whether a session already exists for id.
func (s *Store) Has(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.sessions[id]
	return ok
}

// Todos returns a copy of the session's todos.
func (s *Store) Todos(id string) []model.Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.Todo(nil), s.ensure(id).Todos...)
}

// AddTodo appends a todo and returns it.
func (s *Store) AddTodo(id, title string) model.Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	sess.todoSeq++
	t := model.Todo{ID: sess.todoSeq, Title: title}
	sess.Todos = append(sess.Todos, t)
	return t
}

// ToggleTodo flips Done and returns the updated todo.
func (s *Store) ToggleTodo(id string, todoID int) (model.Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	for i := range sess.Todos {
		if sess.Todos[i].ID == todoID {
			sess.Todos[i].Done = !sess.Todos[i].Done
			return sess.Todos[i], true
		}
	}
	return model.Todo{}, false
}

// UpdateTodo replaces the title and returns the updated todo.
func (s *Store) UpdateTodo(id string, todoID int, title string) (model.Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	for i := range sess.Todos {
		if sess.Todos[i].ID == todoID {
			sess.Todos[i].Title = title
			return sess.Todos[i], true
		}
	}
	return model.Todo{}, false
}

// AddTodoUnique appends a todo unless a todo with the same title already
// exists in the session (case-insensitive). ok is false on conflict.
func (s *Store) AddTodoUnique(id, title string) (model.Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	for _, t := range sess.Todos {
		if strings.EqualFold(t.Title, title) {
			return model.Todo{}, false
		}
	}
	sess.todoSeq++
	t := model.Todo{ID: sess.todoSeq, Title: title}
	sess.Todos = append(sess.Todos, t)
	return t, true
}

// ReplaceTodo fully replaces a todo's fields (PUT semantics, idempotent).
func (s *Store) ReplaceTodo(id string, todoID int, title string, done bool) (model.Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	for i := range sess.Todos {
		if sess.Todos[i].ID == todoID {
			sess.Todos[i].Title = title
			sess.Todos[i].Done = done
			return sess.Todos[i], true
		}
	}
	return model.Todo{}, false
}

// DeleteTodo removes a todo by id.
func (s *Store) DeleteTodo(id string, todoID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	for i := range sess.Todos {
		if sess.Todos[i].ID == todoID {
			sess.Todos = append(sess.Todos[:i], sess.Todos[i+1:]...)
			return true
		}
	}
	return false
}

// Users returns a copy of the session's users.
func (s *Store) Users(id string) []model.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.User(nil), s.ensure(id).Users...)
}

// AddUser appends a user and returns it.
func (s *Store) AddUser(id, name, email string) model.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	sess.userSeq++
	u := model.User{ID: sess.userSeq, Name: name, Email: email}
	sess.Users = append(sess.Users, u)
	return u
}

// UpdateUser replaces name/email and returns the updated user.
func (s *Store) UpdateUser(id string, userID int, name, email string) (model.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	for i := range sess.Users {
		if sess.Users[i].ID == userID {
			sess.Users[i].Name = name
			sess.Users[i].Email = email
			return sess.Users[i], true
		}
	}
	return model.User{}, false
}

// DeleteUser removes a user by id.
func (s *Store) DeleteUser(id string, userID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	for i := range sess.Users {
		if sess.Users[i].ID == userID {
			sess.Users = append(sess.Users[:i], sess.Users[i+1:]...)
			return true
		}
	}
	return false
}

// AppendHistory prepends an exchange (newest first) and trims to maxHistory.
func (s *Store) AppendHistory(id string, ex model.Exchange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.ensure(id)
	sess.History = append([]model.Exchange{ex}, sess.History...)
	if len(sess.History) > maxHistory {
		sess.History = sess.History[:maxHistory]
	}
}

// History returns a copy of the session's exchange log (newest first).
func (s *Store) History(id string) []model.Exchange {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.Exchange(nil), s.ensure(id).History...)
}
