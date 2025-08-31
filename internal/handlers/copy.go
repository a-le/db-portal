package handlers

import (
	"database/sql"
	"db-portal/internal/contextkeys"
	"db-portal/internal/copydata"
	"db-portal/internal/dbutil"
	"db-portal/internal/response"
	"fmt"
	"io"
	"net/http"
	"time"
)

type copyData struct {
	Reads  int `json:"reads"`
	Writes int `json:"writes"`
}

type copyResponse = response.Response[copyData]

func (s *Services) CopyHandler(w http.ResponseWriter, r *http.Request) {
	resp := copyResponse{}

	// Parse multipart form (10 MB max memory, rest to disk)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		resp.Error = "failed to parse multipart form"
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	// Helper to extract endpoint from form
	parseEndpoint := func(prefix string) copydata.EndPoint {
		EPType := r.FormValue(prefix + "[type]")
		switch EPType {
		case "table":
			return copydata.EndPoint{
				Type:       EPType,
				DSName:     r.FormValue(prefix + "[dsName]"),
				Schema:     r.FormValue(prefix + "[schema]"),
				Table:      r.FormValue(prefix + "[table]"),
				IsNewTable: r.FormValue(prefix + "[isNewTable]"),
			}
		case "query":
			return copydata.EndPoint{
				Type:   EPType,
				DSName: r.FormValue(prefix + "[dsName]"),
				Schema: r.FormValue(prefix + "[schema]"),
				Query:  r.FormValue(prefix + "[query]"),
			}
		case "file":
			return copydata.EndPoint{
				Type:   EPType,
				Format: r.FormValue(prefix + "[format]"),
			}
		}
		return copydata.EndPoint{}
	}

	// Retrieve request from form
	var req copydata.CopyRequest
	req.OriginEP = parseEndpoint("origin")
	req.DestEP = parseEndpoint("destination")

	// Retrieve file from form
	var originFile io.Reader
	if req.OriginEP.Type == "file" {
		file, _, err := r.FormFile("origin[file]")
		if err != nil || file == nil {
			resp.Error = "failed to get origin_file"
			response.WriteJSON(w, http.StatusBadRequest, &resp)
			return
		}
		// This copy is an interface assignment: assigning file (multipart.File) to originFile (io.Reader)
		// does NOT copy file data or buffer. Both variables point to the same underlying file stream.
		originFile = file
		defer file.Close()
	}

	// Prepare origin database connection
	var originConn *sql.Conn
	currentUsername := contextkeys.UsernameFromContext(r.Context())
	if req.OriginEP.DSName != "" {
		ds, err := s.Store.RequireUserDataSource(currentUsername, currentUsername, req.OriginEP.DSName)
		if err != nil {
			resp.Error = fmt.Sprintf("origin data source %v not found or not allowed", req.OriginEP.DSName)
			response.WriteJSON(w, http.StatusNotFound, &resp)
			return
		}
		if originConn, err = dbutil.GetConn(r.Context(), ds.Vendor, ds.Location, false); err != nil {
			resp.Error = fmt.Sprintf("failed to connect to origin: %v", err)
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
		req.OriginEP.DBVendor = ds.Vendor

		// set schema
		schema := req.OriginEP.Schema
		if schema != "" {
			setSchema, args, err := s.CommandsConfig.Data.Command("set-schema", req.OriginEP.DBVendor, []string{schema})
			if err != nil {
				resp.Error = err.Error()
				response.WriteJSON(w, http.StatusInternalServerError, &resp)
				return
			}
			if _, err = originConn.ExecContext(r.Context(), setSchema, args...); err != nil {
				resp.Error = err.Error()
				response.WriteJSON(w, http.StatusInternalServerError, &resp)
				return
			}
		}
	}

	// Create src row reader
	src, err := copydata.NewRowReader(req.OriginEP, r.Context(), originConn, originFile)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	// Prepare destination database transaction
	var destTx *sql.Tx
	if req.DestEP.Type == "table" {
		var destConn *sql.Conn
		ds, err := s.Store.RequireUserDataSource(currentUsername, currentUsername, req.DestEP.DSName)
		if err != nil {
			resp.Error = fmt.Sprintf("destination data source %v not found or not allowed", req.DestEP.DSName)
			response.WriteJSON(w, http.StatusNotFound, &resp)
			return
		}
		if destConn, err = dbutil.GetConn(r.Context(), ds.Vendor, ds.Location, false); err != nil {
			resp.Error = fmt.Sprintf("failed to connect to destination: %v", err)
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
		req.DestEP.DBVendor = ds.Vendor

		// set schema
		schema := req.DestEP.Schema
		if schema != "" {
			setSchema, args, err := s.CommandsConfig.Data.Command("set-schema", req.DestEP.DBVendor, []string{schema})
			if err != nil {
				resp.Error = err.Error()
				response.WriteJSON(w, http.StatusInternalServerError, &resp)
				return
			}
			if _, err = destConn.ExecContext(r.Context(), setSchema, args...); err != nil {
				resp.Error = err.Error()
				response.WriteJSON(w, http.StatusInternalServerError, &resp)
				return
			}
		}

		// Start transaction
		destTx, err = destConn.BeginTx(r.Context(), nil)
		if err != nil {
			resp.Error = fmt.Sprintf("failed to begin transaction: %v", err)
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}

		defer destTx.Rollback() // Safe to call even if already committed
	}

	// Prepare destination file for streaming.
	var destWriter io.Writer
	if req.DestEP.Type == "file" {
		destWriter = w // use http.ResponseWriter as destWriter to stream file to client
		ext := req.DestEP.Format
		if len(ext) > 4 {
			ext = req.DestEP.Format[:4]
		}
		w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+"."+ext)
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Create dest row writer
	dst, err := copydata.NewRowWriter(req.DestEP, r.Context(), destTx, destWriter, src.Fields())
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	// Copy data
	resp.Data.Reads, resp.Data.Writes, err = copydata.CopyData(src, dst)

	// Handle transaction commit/rollback for database destination
	if req.DestEP.Type == "table" {
		if err != nil {
			fmt.Println("rollback")
			destTx.Rollback()
		} else {
			if commitErr := destTx.Commit(); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}

	// End file response
	if req.DestEP.Type == "file" {
		if err != nil {
			if resp.Data.Writes == 0 {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				// error message is appended to end of file
				w.Write([]byte(err.Error()))
			}
		}
		return
	}

	// send response
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}
	response.WriteJSON(w, http.StatusOK, &resp)
}
