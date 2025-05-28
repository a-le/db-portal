package handlers

import (
	"net/http"
)

func (s *Services) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// nothing to do
	// Is meant to be used with bad credentials so that the browser forgets those credentials
}
