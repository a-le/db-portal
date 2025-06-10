package handlers

import (
	"net/http"
	"time"
)

// LogoutHandler handles user logout by clearing the JWT cookie and returning a 401 status
// to trigger a client-side redirect to the login page.
func (s *Services) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Delete the JWT cookie by setting it to expire in the past
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-24 * time.Hour), // Set expiration in the past
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	// Return 401 to trigger the client-side redirect
	http.Error(w, "Logged out", http.StatusUnauthorized)
}
