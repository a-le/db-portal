package handlers

import (
	"db-portal/internal/db"
	"db-portal/internal/security"
	"fmt"
	"net/http"
)

func (s *Services) ExportHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(security.UserContextKey).(string)
	conname := r.FormValue("conn")
	exportType := r.FormValue("exportType")
	gz := r.FormValue("gz")
	query := r.FormValue("query")
	s.CommandsConfig.Reload()

	// get connection details
	connDetails, err := s.Store.FetchConn(username, conname)
	if err != nil {
		http.Error(w, fmt.Sprintf("connection <%v> not found or not allowed for user <%v>", conname, username), http.StatusNotFound)
		return
	}

	// get DB connection
	conn, err := db.GetConn(connDetails.DBVendor, connDetails.DSN, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Set schema if provided
	schema := r.FormValue("schema")
	if schema != "" {
		setSchema, args, _ := s.CommandsConfig.Data.Command("set-schema", connDetails.DBVendor, []string{schema})
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
		if err := s.Exporter.ExportCSV(w, rows, gz == "on"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "json":
		if err := s.Exporter.ExportJSON(w, rows, gz == "on"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "jsontabular":
		s.Exporter.SetDBVendor(connDetails.DBVendor)
		if err := s.Exporter.ExportJSONTabular(w, rows, gz == "on"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "xlsx":
		if err := s.Exporter.ExportXLSX(w, rows, gz == "on"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Export type not supported", http.StatusBadRequest)
	}
}
