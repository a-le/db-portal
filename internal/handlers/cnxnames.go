package handlers

import (
	"db-portal/internal/response"
	"db-portal/internal/security"
	"net/http"
)

func (s *Services) CnxNamesHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(security.UserContextKey).(string)

	rows, err := s.Store.FetchUserConns(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.SendJSON(&response.Data{
		Data: rows,
	}, w)
}
