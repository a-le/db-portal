package handlers

import (
	"db-portal/internal/auth"
	"db-portal/internal/response"
	"net/http"
)

func (s *Services) CnxNamesHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(auth.UserContextKey).(string)

	rows, err := s.Store.FetchUserConns(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.SendJSON(&response.Data{
		Data: rows,
	}, w)
}
