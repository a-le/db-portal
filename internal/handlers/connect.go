package handlers

import (
	"db-portal/internal/db"
	"db-portal/internal/internaldb"
	"db-portal/internal/response"
	"db-portal/internal/security"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Services) ConnectHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(security.UserContextKey).(string) // Retrieve the username from the context
	conname := chi.URLParam(r, "conn")
	var err error

	// get connection details
	var connDetails internaldb.ConnDetails
	if connDetails, err = s.Store.FetchConn(username, conname); err != nil {
		http.Error(w, fmt.Sprintf("connection %v not found or not allowed", conname), http.StatusNotFound)
		return
	}

	// try to get conn from DB server
	var conn db.Conn
	var dResult db.DResult
	if conn, dResult.DBerror = db.GetConn(connDetails.DBVendor, connDetails.DSN, true); dResult.DBerror == nil {
		conn.Close()
	}

	// send response
	response.SendJSON(&dResult, w)
}
