package handlers

import (
	"db-portal/internal/auth"
	"db-portal/internal/db"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/sqltocsv"
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
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+".csv")
		if err := sqltocsv.Write(w, rows); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "xlsx":
		tmpfile, err := os.CreateTemp("", "dbexport_*.xlsx")
		if err != nil {
			http.Error(w, "unable to create temp file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmpfile.Name())
		defer tmpfile.Close()

		if err := db.RowsToXlsx(rows, tmpfile.Name()); err != nil {
			http.Error(w, "failed to generate XLSX: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+".xlsx")

		if _, err := io.Copy(w, tmpfile); err != nil {
			http.Error(w, "Unable to copy file", http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(w, tmpfile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Export type not supported", http.StatusBadRequest)
	}
}
