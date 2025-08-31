package handlers

import (
	"database/sql"
	"db-portal/internal/dbutil"
	"db-portal/internal/response"

	"db-portal/internal/contextkeys"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type commandResp = response.Response[dbutil.DBResult]

// commands are defined in config/commands.jsonc
// some commands do mot exists for some drivers
func (s *Services) CommandHandler(w http.ResponseWriter, r *http.Request) {

	resp := commandResp{}

	currentUsername := contextkeys.UsernameFromContext(r.Context())
	dsName := chi.URLParam(r, "dsName")
	schema := chi.URLParam(r, "schema")
	commandName := chi.URLParam(r, "command")

	// reload config files if needed
	s.CommandsConfig.Reload()

	// get ds info from internal DB
	ds, err := s.Store.RequireUserDataSource(currentUsername, currentUsername, dsName)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	// get DB connection
	var conn *sql.Conn
	conn, err = dbutil.GetConn(r.Context(), ds.Vendor, ds.Location, true)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusOK, &resp)
		return
	}
	defer conn.Close()

	// Set the schema if specified.
	// Then, defer a query to restore the schema to the default before the connection is returned to the DB pool.
	if schema != "" {
		setSchema, args, _ := s.CommandsConfig.Data.Command("set-schema", ds.Vendor, []string{schema})
		setSchemaDefault, _, _ := s.CommandsConfig.Data.Command("set-schema-default", ds.Vendor, []string{})
		if setSchemaDefault == "" {
			resp.Error = fmt.Sprintf("a 'set schema' command was defined, but the 'set-schema-default' command is empty. Driver is %s\n", ds.Vendor)
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
		if _, err = conn.ExecContext(r.Context(), setSchema, args...); err != nil {
			resp.Data.DBerror = err.Error()
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
		defer conn.ExecContext(r.Context(), setSchemaDefault, []any{}...)
	}

	// get list of args from the query string. Those are SQL identifiers for the SQL command
	urlArgs := r.URL.Query()["args"]

	// fetch command and args
	var command string
	var args []any

	if command, args, err = s.CommandsConfig.Data.Command(commandName, ds.Vendor, urlArgs); err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	// send response if command is not implemented
	if command == "" {
		resp.Data.DBerror = fmt.Sprintf("command <%v> is not supported for %v", commandName, ds.Vendor)
		resp.Error = fmt.Sprintf("command <%v> is not supported for %v", commandName, ds.Vendor)
		response.WriteJSON(w, http.StatusOK, &resp)
		return
	}

	// run the command
	if resp.Data, err = dbutil.QueryWithResult(r.Context(), conn, command, args, 0); err != nil {
		resp.Data.DBerror = err.Error()
		response.WriteJSON(w, http.StatusOK, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}
