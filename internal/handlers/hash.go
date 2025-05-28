package handlers

import (
	"db-portal/internal/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// return a bcrypt hash of a string (useful for password hashing)
// there is some salt in the hash, so the result will be different each time
func (s *Services) HashHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(auth.HashPassword(chi.URLParam(r, "string"))))
}
