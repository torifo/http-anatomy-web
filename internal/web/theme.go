package web

import "net/http"

const themeCookie = "ha_theme"

// themeFrom returns the visitor's theme ("dark" default) from the cookie.
func themeFrom(r *http.Request) string {
	if c, err := r.Cookie(themeCookie); err == nil && c.Value == "light" {
		return "light"
	}
	return "dark"
}

// handleThemeToggle flips the theme cookie and asks htmx to reload via the
// HX-Redirect response header — demonstrating a header-driven client
// redirect plus server-rendered (cookie-based) UI state, no client JS state.
func (s *Server) handleThemeToggle(w http.ResponseWriter, r *http.Request) {
	next := "light"
	if themeFrom(r) == "light" {
		next = "dark"
	}
	http.SetCookie(w, &http.Cookie{
		Name:     themeCookie,
		Value:    next,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
