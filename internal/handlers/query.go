package handlers

import (
	"context"
	"db-portal/internal/contextkeys"
	"db-portal/internal/dbutil"
	"db-portal/internal/response"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type queryResp = response.Response[dbutil.DBResult]

func (s *Services) QueryHandler(w http.ResponseWriter, r *http.Request) {

	resp := queryResp{}

	currentUsername := contextkeys.UsernameFromContext(r.Context())
	dsName := chi.URLParam(r, "dsName")
	schema := chi.URLParam(r, "schema")

	// reload config files if needed
	s.CommandsConfig.Reload()

	// get ds info from internal DB
	ds, err := s.Store.RequireUserDataSource(currentUsername, currentUsername, dsName)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	// get conn
	conn, err := dbutil.GetConn(ds.Vendor, ds.Location, false)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}
	defer conn.Close()

	// set schema
	if schema != "" {
		setSchema, args, err := s.CommandsConfig.Data.Command("set-schema", ds.Vendor, []string{schema})
		if err != nil {
			resp.Error = err.Error()
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
		if _, err = conn.ExecContext(r.Context(), setSchema, args...); err != nil {
			resp.Error = err.Error()
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
	}

	query := r.FormValue("query")

	// Build explain query
	if r.FormValue("explain") == "1" {
		command, _, err := s.CommandsConfig.Data.Command("explain", ds.Vendor, []string{})
		if err != nil {
			resp.Error = err.Error()
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
		if command == "" {
			resp.Data.DBerror = fmt.Sprintf("explain command is not supported for the %v database", ds.Vendor)
			response.WriteJSON(w, http.StatusOK, &resp)
			return
		}

		if ds.Vendor == "mssql" {
			// explain command (mssql SET SHOWPLAN_ALL ON) is executed before the query
			if _, err = conn.ExecContext(r.Context(), command, []any{}...); err != nil {
				resp.Error = err.Error()
				response.WriteJSON(w, http.StatusInternalServerError, &resp)
				return
			}
		} else {
			query = command + " " + query
		}
	}

	// infer statement type (query or not query) and command (select, insert, update, delete, etc.)
	stmtInfos := dbutil.StmtInfo(query, ds.Vendor)

	// execute query
	ctx := r.Context()
	if stmtInfos.Type == "query" {
		resp.Data, err = dbutil.QueryWithResult(ctx, conn, query, []any{}, int64(s.ServerConfig.Data.MaxResultsetLength))
	} else {
		resp.Data, err = dbutil.ExecWithResult(ctx, conn, query, []any{})
	}
	resp.Data.StmtType = stmtInfos.Type
	resp.Data.StmtCmd = stmtInfos.Cmd

	if err != nil {
		// just set error. http.StatusOK is fine here
		if ctx.Err() == context.Canceled {
			resp.Error = "request canceled by client"
		}
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}
