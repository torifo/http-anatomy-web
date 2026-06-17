package web

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const sessionCookie = "ha_session"

// newSessionID returns a random hex session id.
func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// sessionID returns the caller's session id from the cookie, issuing a new
// one (and Set-Cookie) when absent. The cookie is HttpOnly + SameSite=Lax;
// it carries no sensitive data, only an opaque identifier.
func sessionID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
		return c.Value
	}
	id := newSessionID()
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}
