// Package web wires HTTP routing, session cookies, and the HTMX handlers
// that return a primary fragment plus an out-of-band inspector update.
package web

import (
	"net/http"

	"http-anatomy/internal/store"
)

// Server holds dependencies shared by the handlers.
type Server struct {
	store *store.Store
}

// NewServer builds the HTTP handler with all routes registered.
func NewServer(s *store.Store) http.Handler {
	srv := &Server{store: s}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", srv.handleIndex)
	mux.HandleFunc("GET /theme/toggle", srv.handleThemeToggle)
	mux.HandleFunc("GET /fragments/todos", srv.handleTodosFragment)
	mux.HandleFunc("GET /fragments/users", srv.handleUsersFragment)
	mux.HandleFunc("POST /api/todos", srv.handleCreateTodo)
	mux.HandleFunc("PUT /api/todos/{id}", srv.handlePutTodo)
	mux.HandleFunc("PATCH /api/todos/{id}", srv.handlePatchTodo)
	mux.HandleFunc("DELETE /api/todos/{id}", srv.handleDeleteTodo)
	mux.HandleFunc("POST /api/users", srv.handleCreateUser)
	mux.HandleFunc("PATCH /api/users/{id}", srv.handlePatchUser)
	mux.HandleFunc("DELETE /api/users/{id}", srv.handleDeleteUser)
	return mux
}
