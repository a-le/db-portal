package handlers

import (
	"context"
	"db-portal/internal/auth"
	"db-portal/internal/db"
	"db-portal/internal/response"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// commands are defined in config/commands.jsonc
// some commands do mot exists for some drivers
func (s *Services) CommandHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(auth.UserContextKey).(string)
	conname := chi.URLParam(r, "conn")
	schema := chi.URLParam(r, "schema")
	commandName := chi.URLParam(r, "command")

	// reload config files if needed
	s.CommandsConfig.Reload()

	// get connection details
	connDetails, err := s.Store.FetchConn(username, conname)
	if err != nil {
		http.Error(w, fmt.Sprintf("connection %v not found or not allowed", conname), http.StatusNotFound)
		return
	}

	var dResult db.DResult

	// get DB connection
	var conn db.Conn
	conn, dResult.DBerror = db.GetConn(connDetails.DBType, connDetails.DSN, true)
	if dResult.DBerror != nil {
		response.SendJSON(&dResult, w)
		return
	}
	defer conn.Close()

	// Set the schema if specified.
	// Then, defer a query to restore the schema to the default before the connection is returned to the DB pool.
	if schema != "" {
		setSchema, args, _ := s.CommandsConfig.Data.Command("set-schema", connDetails.DBType, []string{schema})
		setSchemaDefault, _, _ := s.CommandsConfig.Data.Command("set-schema-default", connDetails.DBType, []string{})
		if setSchemaDefault == "" {
			http.Error(w, fmt.Sprintf("a 'set schema' command was defined, but the 'set-schema-default' command is empty. Driver is %s\n", connDetails.DBType), http.StatusInternalServerError)
			return
		}
		if _, err = db.ExecContext(conn, setSchema, args); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.ExecContext(conn, setSchemaDefault, []any{})
	}

	// get list of args from the query string. Those are SQL identifiers for the SQL command
	var urlArgs []string
	for i := 0; ; i++ {
		v := r.URL.Query().Get(fmt.Sprintf("args[%d]", i)) // Get args from the query string (indexed parameters, ex: ?args[0]=foo&args[1]=bar)
		if v == "" {
			break
		}
		urlArgs = append(urlArgs, v)
	}

	// fetch command and args
	var command string
	var args []any

	if command, args, err = s.CommandsConfig.Data.Command(commandName, connDetails.DBType, urlArgs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send response if command is not implemented
	if command == "" {
		dResult.DBerror = fmt.Errorf("command <%v> is not supported for %v", commandName, connDetails.DBType)
		response.SendJSON(&dResult, w)
		return
	}

	// run the command
	if dResult, err = db.DQueryContext(context.Background(), conn, command, args, 0); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.SendJSON(&dResult, w)
}
