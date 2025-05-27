package handlers

import (
	"context"
	"db-portal/internal/auth"
	"db-portal/internal/db"
	"db-portal/internal/response"
	"fmt"
	"net/http"
)

func (s *Services) QueryHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(auth.UserContextKey).(string)
	conname := r.FormValue("conn")

	// reload config files if needed
	s.CommandsConfig.Reload()

	// get connection details
	connDetails, err := s.Store.FetchConn(username, conname)
	if err != nil {
		http.Error(w, fmt.Sprintf("connection %v not found or not allowed", conname), http.StatusNotFound)
		return
	}

	// get conn
	conn, err := db.GetConn(connDetails.DBType, connDetails.DSN, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// set schema
	if schema := r.FormValue("schema"); schema != "" {
		setSchema, args, _ := s.CommandsConfig.Data.Command("set-schema", connDetails.DBType, []string{schema})
		if _, err = db.ExecContext(conn, setSchema, args); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	query := r.FormValue("query")

	// Build explain query
	if r.FormValue("explain") == "1" {
		command, _, err := s.CommandsConfig.Data.Command("explain", connDetails.DBType, []string{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if command == "" {
			dResult := db.DResult{}
			dResult.DBerror = fmt.Errorf("explain command is not supported for the %v database", connDetails.DBType)
			response.SendJSON(&dResult, w)
			return
		}

		if connDetails.DBType == "mssql" {
			_, err = db.ExecContext(conn, command, []any{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			query = command + " " + query
		}
	}

	// infer statement type (query or not query) and command (select, insert, update, delete, etc.)
	stmtInfos := db.StmtInfos(query, connDetails.DBType)

	// execute query
	var dResult db.DResult
	ctx := r.Context()
	if stmtInfos.Type == "query" {
		dResult, err = db.DQueryContext(ctx, conn, query, []any{}, int64(s.ServerConfig.Data.MaxResultsetLength))
	} else {
		dResult, err = db.DExecContext(ctx, conn, query, []any{})
	}
	dResult.StmtType = stmtInfos.Type
	dResult.StmtCmd = stmtInfos.Cmd

	if err != nil {
		if ctx.Err() == context.Canceled {
			http.Error(w, "request canceled by client", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.SendJSON(&dResult, w)
}
