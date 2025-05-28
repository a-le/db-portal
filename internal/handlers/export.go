package handlers

import (
	"db-portal/internal/auth"
	"db-portal/internal/db"
	"fmt"
	"net/http"
)

func (s *Services) ExportHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(auth.UserContextKey).(string)
	conname := r.FormValue("conn")
	exportType := r.FormValue("exportType")
	query := r.FormValue("query")
	s.CommandsConfig.Reload()

	// get connection details
	connDetails, err := s.Store.FetchConn(username, conname)
	if err != nil {
		http.Error(w, fmt.Sprintf("connection %v not found or not allowed", conname), http.StatusNotFound)
		return
	}

	// get DB connection
	conn, err := db.GetConn(connDetails.DBType, connDetails.DSN, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Set schema if provided
	schema := r.FormValue("schema")
	if schema != "" {
		setSchema, args, _ := s.CommandsConfig.Data.Command("set-schema", connDetails.DBType, []string{schema})
		if _, err := db.ExecContext(conn, setSchema, args); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Run the query
	rows, err := db.QueryContext(r.Context(), conn, query, []any{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	switch exportType {
	case "csv":
		if err := s.Exporter.ExportCSV(w, rows); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "json":
		if err := s.Exporter.ExportJSON(w, rows); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "xlsx":
		if err := s.Exporter.ExportXLSX(w, rows); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Export type not supported", http.StatusBadRequest)
	}
}
