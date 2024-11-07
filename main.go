// Copyright (C) 2024 https://github.com/a-le
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"godatabaseadmin/internal/auth"
	"godatabaseadmin/internal/config"
	"godatabaseadmin/internal/db"
	"godatabaseadmin/internal/jsminifier"
	"godatabaseadmin/internal/response"
	"godatabaseadmin/internal/timer"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/sqltocsv"
)

const Version = "1.0.0-b1"
const AppName = "goDatabaseAdmin"

func main() {

	/* set config dir */
	var err error
	var configDir string
	if len(os.Args) > 1 {
		configDir = os.Args[1]
	} else {
		executablePath, err := os.Executable()
		if err != nil {
			panic(fmt.Sprintf("error getting executable path: %v", err))
		}
		configDir = filepath.Dir(executablePath) + "/conf"
	}

	/* config */
	serverConfig := config.New[config.Server](configDir + "/server.yaml")
	usersConfig := config.New[config.UsersConfig](configDir + "/users.yaml")
	connectionsConfig := config.New[config.ConnectionsConfig](configDir + "/connections.yaml")
	commandsConfig := config.New[config.CommandsConfig](configDir + "/commands.yaml")

	/* load config files */
	if err := serverConfig.Load(); err != nil {
		panic(fmt.Sprintf("error loading %v file: %v", serverConfig.Filename, err))
	}
	if err := usersConfig.Load(); err != nil {
		panic(fmt.Sprintf("error loading %v file: %v", usersConfig.Filename, err))
	}
	if err := connectionsConfig.Load(); err != nil {
		panic(fmt.Sprintf("error loading %v file: %v", connectionsConfig.Filename, err))
	}
	if err := commandsConfig.Load(); err != nil {
		panic(fmt.Sprintf("error loading %v file: %v", commandsConfig.Filename, err))
	}

	/* auth variables */
	htpasswdPath := serverConfig.Data.HtpasswdFile
	jwtSecretKey := serverConfig.Data.JWTSecretKey

	/* chi router */
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5, "text/html", "text/css", "application/json", "text/javascript"))

	// connect endpoint
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Get("/api/connect/{conn}", func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(auth.UserContextKey).(string) // Retrieve the username from the context
		conname := chi.URLParam(r, "conn")

		// reload config files if needed
		usersConfig.Reload()
		connectionsConfig.Reload()

		// check connection validity for this user
		if !slices.Contains(usersConfig.Data[username].Connections, conname) {
			http.Error(w, fmt.Sprintf("connection %v not allowed", conname), http.StatusForbidden)
			return
		}

		// get connection config
		connConfig, ok := connectionsConfig.Data[conname]
		if !ok {
			http.Error(w, fmt.Sprintf("connection %v not found", conname), http.StatusNotFound)
			return
		}

		// try to get conn from DB server
		qResult := db.NewQResult()
		var conn db.Conn
		conn, qResult.DBerror = db.GetConn(connConfig.DBType, connConfig.DSN, true)
		if qResult.DBerror == nil {
			conn.Close()
		}

		// send response
		response.SendJSON(&qResult, w)
	})

	// export endpoint
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Post("/api/export", func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(auth.UserContextKey).(string) // Retrieve the username from the context
		conname := r.FormValue("conn")

		// reload config files if needed
		usersConfig.Reload()
		connectionsConfig.Reload()
		commandsConfig.Reload()

		// check connection validity for this user
		if !slices.Contains(usersConfig.Data[username].Connections, conname) {
			http.Error(w, fmt.Sprintf("connection %v not allowed", conname), http.StatusForbidden)
			return
		}

		// get connection
		connConfig, ok := connectionsConfig.Data[conname]
		if !ok {
			http.Error(w, fmt.Sprintf("connection %v not found", conname), http.StatusNotFound)
			return
		}

		// get conn
		var conn db.Conn
		var err error
		conn, err = db.GetConn(connConfig.DBType, connConfig.DSN, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// set schema if set
		schema := r.FormValue("schema")
		if schema != "" {
			setSchema, args, _ := commandsConfig.Data.Command("set-schema", connConfig.DBType, []string{schema})
			if setSchema != "" {
				_, err = db.Exec(conn, setSchema, args)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// export
		exportType := r.FormValue("exportType")
		query := r.FormValue("query")
		ctx := r.Context()

		// csv export: execute the query and send csv file
		if exportType == "csv" {
			args := []any{}
			var rows *sql.Rows
			if rows, err = db.Query(ctx, conn, query, args); err != nil {
				fmt.Printf("err: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+".csv")

			if err = sqltocsv.Write(w, rows); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		// xlsx export: execute the query and send .xlsx file
		if exportType == "xlsx" {
			args := []any{}
			var rows *sql.Rows
			if rows, err = db.Query(ctx, conn, query, args); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			tempFile, err := os.CreateTemp("", "dbexport_*.tmp")
			if err != nil {
				http.Error(w, "Unable to create temporary file", http.StatusInternalServerError)
				return
			}
			defer os.Remove(tempFile.Name())

			err = db.RowsToXlsx(rows, tempFile.Name())
			if err != nil {
				http.Error(w, "Failed to generate XLSX: "+err.Error(), http.StatusInternalServerError)
			}

			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+".xlsx")

			if _, err := io.Copy(w, tempFile); err != nil {
				http.Error(w, "Unable to stream file", http.StatusInternalServerError)
				return
			}

			if err := tempFile.Close(); err != nil {
				http.Error(w, "Unable to close temporary file", http.StatusInternalServerError)
				return
			}

			return
		}

		http.Error(w, "Export type not supported", http.StatusInternalServerError)
	})

	// query endpoint
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Post("/api/query", func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(auth.UserContextKey).(string) // Retrieve the username from the context
		conname := r.FormValue("conn")

		// reload config files if needed
		usersConfig.Reload()
		connectionsConfig.Reload()
		commandsConfig.Reload()

		// check connection validity for this user
		if !slices.Contains(usersConfig.Data[username].Connections, conname) {
			http.Error(w, fmt.Sprintf("connection %v not allowed", conname), http.StatusForbidden)
			return
		}

		// get connection
		connConfig, ok := connectionsConfig.Data[conname]
		if !ok {
			http.Error(w, fmt.Sprintf("connection %v not found", conname), http.StatusNotFound)
			return
		}

		query := r.FormValue("query")

		// get conn
		var conn db.Conn
		var err error
		conn, err = db.GetConn(connConfig.DBType, connConfig.DSN, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// set schema if set
		schema := r.FormValue("schema")
		if schema != "" {
			setSchema, args, _ := commandsConfig.Data.Command("set-schema", connConfig.DBType, []string{schema})
			if setSchema != "" {
				_, err = db.Exec(conn, setSchema, args)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// Build explain query
		if r.FormValue("explain") == "1" {
			command, _, err := commandsConfig.Data.Command("explain", connConfig.DBType, []string{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if command == "" {
				qResult := db.NewQResult()
				qResult.DBerror = fmt.Errorf("explain command is not supported for the %v database", connConfig.DBType)
				response.SendJSON(&qResult, w)
				return
			}

			if connConfig.DBType == "mssql" {
				_, err = db.Exec(conn, command, []any{})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				query = command + " " + query
			}
		}

		// execute the query and send json
		ctx := r.Context()
		qResult := db.NewQResult()
		args := []any{}
		qResult, err = db.QQuery(ctx, conn, query, args, serverConfig.Data.MaxResultsetLength, r.FormValue("statementType"))
		if err != nil {
			if ctx.Err() == context.Canceled {
				//fmt.Println("request canceled by client")
				http.Error(w, "request canceled by client", http.StatusRequestTimeout)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response.SendJSON(&qResult, w)
	})

	// command endpoint. A command is a SQL statement for the UI
	// commands are defined in config/commands.jsonc
	// some commands do mot exists for some drivers
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Get("/api/command/{conn}/{schema}/{command}", func(w http.ResponseWriter, r *http.Request) {
		var err error

		// Retrieve the username from the context
		username := r.Context().Value(auth.UserContextKey).(string)
		conname := chi.URLParam(r, "conn")

		// reload config files if needed
		usersConfig.Reload()
		connectionsConfig.Reload()
		commandsConfig.Reload()

		// check connection validity for this user
		if !slices.Contains(usersConfig.Data[username].Connections, conname) {
			http.Error(w, fmt.Sprintf("connection <%v> is not allowed", conname), http.StatusForbidden)
			return
		}

		// get connection from config
		connConfig, ok := connectionsConfig.Data[conname]
		if !ok {
			http.Error(w, fmt.Sprintf("connection <%v> could not be found", conname), http.StatusNotFound)
			return
		}

		qResult := db.NewQResult()

		// get list of args from the query string. Those are SQL identifiers for the SQL command
		var urlArgs []string
		for i := 0; ; i++ {
			v := r.URL.Query().Get(fmt.Sprintf("args[%d]", i)) // Get args from the query string (indexed parameters, ex: ?args[0]=foo&args[1]=bar)
			if v == "" {
				break
			}
			urlArgs = append(urlArgs, v)
		}
		// build SQL command
		command, args, err := commandsConfig.Data.Command(chi.URLParam(r, "command"), connConfig.DBType, urlArgs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// send response if command is not implemented
		if command == "" {
			qResult.DBerror = fmt.Errorf("command <%v> is not supported for %v", chi.URLParam(r, "command"), connConfig.DBType)
			response.SendJSON(&qResult, w)
			return
		}

		// conn
		var conn db.Conn
		conn, qResult.DBerror = db.GetConn(connConfig.DBType, connConfig.DSN, true)
		if qResult.DBerror != nil {
			response.SendJSON(&qResult, w)
			return
		}
		defer conn.Close()

		// Set the schema if specified.
		// Then, defer a query to restore the schema to the default before the connection is returned to the DB pool.
		// This is necessary because a connection pool is used for internal queries, and the schema change will persist.
		schema := chi.URLParam(r, "schema")
		if schema != "" {
			setSchema, args, _ := commandsConfig.Data.Command("set-schema", connConfig.DBType, []string{schema})
			setSchemaDefault, _, _ := commandsConfig.Data.Command("set-schema-default", connConfig.DBType, []string{})
			if setSchema != "" {
				if setSchemaDefault != "" {
					_, err = db.Exec(conn, setSchema, args)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					defer db.Exec(conn, setSchemaDefault, []any{})
				} else {
					http.Error(w, fmt.Sprintf("a 'set schema' command was defined, but the 'set-schema-default' command is empty. Driver is %s\n", connConfig.DBType), http.StatusInternalServerError)
					return
				}

			}
		}

		// run the command
		if command != "" {
			if qResult, err = db.QQuery(context.Background(), conn, command, args, 0, "query"); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		response.SendJSON(&qResult, w)
	})

	// cnxnames endpoint
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Get("/api/config/cnxnames", func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(auth.UserContextKey).(string) // Retrieve the username from the context

		// reload config files if needed
		usersConfig.Reload()
		connectionsConfig.Reload()

		var rows [][]string
		for _, cnxname := range usersConfig.Data[username].Connections {
			config, exists := connectionsConfig.Data[cnxname]
			if !exists {
				fmt.Printf("connection %v for user %v not found.\n", cnxname, username)
				continue
			}
			rows = append(rows, []string{cnxname, config.DBType})
		}

		response.SendJSON(&response.Data{
			Data: rows,
		}, w)
	})

	// clockresolution endpoint
	var clockResolution time.Duration
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Get("/api/clockresolution", func(w http.ResponseWriter, r *http.Request) {
		if clockResolution == 0 {
			clockResolution = timer.EstimateMinClockResolution(10000)
		}

		response.SendJSON(&response.Data{
			Data: clockResolution,
		}, w)
	})

	// return a bcrypt hash of a string (useful for password hashing)
	// there is some salt in the hash, so the result will be different each time
	r.Get("/hash/{string}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(auth.HashPassword(chi.URLParam(r, "string"))))
	})

	// logout endpoint. Is meant to be used with bad credentials so that the browser forgets those credentials
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		// nothing to do
	})

	// serve static files
	fileServer := http.FileServer(http.Dir("./web"))
	r.Get("/web/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/web", fileServer).ServeHTTP(w, r)
	})

	// index page
	r.With(auth.Auth(htpasswdPath, jwtSecretKey)).Get("/", func(w http.ResponseWriter, r *http.Request) {
		// check if min.js needs update
		jsPath := "./web/cmp"
		minjsPath := "./web/main.min.js"
		jsInfos, err := jsminifier.GetInfos(jsPath, minjsPath)
		if err != nil {
			fmt.Println("error checking if the JS minified version needs updating", err)
		}
		if jsInfos.Expired {
			if err = jsminifier.Combinify(jsPath, minjsPath); err != nil {
				fmt.Println("error while minifying JS", err)
			} else {
				fmt.Println(minjsPath + " has been updated; a new minified version has been created")
			}
		}

		// some js injected in the index.html
		cssInfo, _ := os.Stat("./web/style.css")
		jsCode := `<script>const versionInfo = { js: '%d', css: '%d', server: '%s', appName: '%s' };const username = '%s';</script>`
		js := fmt.Sprintf(jsCode, jsInfos.ModTime().Unix(), cssInfo.ModTime().Unix(), Version, AppName, r.Context().Value(auth.UserContextKey).(string))

		var data []byte
		if data, err = os.ReadFile("./web/index.html"); err != nil {
			fmt.Println("error while reading index.html")
		}
		html := strings.Replace(string(data), "{{.js}}", js, 1)

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	http := &http.Server{
		Addr:    serverConfig.Data.Addr,
		Handler: r,
	}

	if serverConfig.Data.CertFile != "" && serverConfig.Data.KeyFile != "" {
		fmt.Printf("HTTPS server is listening on %s\n", serverConfig.Data.Addr)
		err = http.ListenAndServeTLS(serverConfig.Data.CertFile, serverConfig.Data.KeyFile)
	} else {
		fmt.Printf("HTTP server is listening on %s\n", serverConfig.Data.Addr)
		err = http.ListenAndServe()
	}

	if err != nil {
		panic(fmt.Sprintf("main: listening on %s failed: %v", http.Addr, err))
	}

}
