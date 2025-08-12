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
	parseEndpoint := func(prefix string) copydata.DataEndpoint {
		switch r.FormValue(prefix + "[type]") {
		case "table":
			return copydata.DataEndpoint{
				Type:   r.FormValue(prefix + "[type]"),
				DSName: r.FormValue(prefix + "[dsName]"),
				Schema: r.FormValue(prefix + "[schema]"),
				Table:  r.FormValue(prefix + "[table]"),
			}
		case "query":
			return copydata.DataEndpoint{
				Type:   r.FormValue(prefix + "[type]"),
				DSName: r.FormValue(prefix + "[dsName]"),
				Schema: r.FormValue(prefix + "[schema]"),
				Query:  r.FormValue(prefix + "[query]"),
			}
		case "file":
			return copydata.DataEndpoint{
				Type:   r.FormValue(prefix + "[type]"),
				Format: r.FormValue(prefix + "[format]"),
			}
		}
		return copydata.DataEndpoint{}
	}

	// Retrieve request from form
	var req copydata.DataTransferRequest
	req.Origin = parseEndpoint("origin")
	req.Destination = parseEndpoint("destination")

	// Retrieve file from form
	var originFile io.Reader
	if req.Origin.Type == "file" {
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
	if req.Origin.DSName != "" {
		ds, err := s.Store.RequireUserDataSource(currentUsername, currentUsername, req.Origin.DSName)
		if err != nil {
			resp.Error = fmt.Sprintf("origin data source %v not found or not allowed", req.Origin.DSName)
			response.WriteJSON(w, http.StatusNotFound, &resp)
			return
		}
		if originConn, err = dbutil.GetConn(ds.Vendor, ds.Location, false); err != nil {
			resp.Error = fmt.Sprintf("failed to connect to origin: %v", err)
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
	}

	// Prepare destination database connection
	var destConn *sql.Conn
	var dbVendor string
	if req.Destination.DSName != "" {
		ds, err := s.Store.RequireUserDataSource(currentUsername, currentUsername, req.Destination.DSName)
		dbVendor = ds.Vendor
		if err != nil {
			resp.Error = fmt.Sprintf("destination data source %v not found or not allowed", req.Destination.DSName)
			response.WriteJSON(w, http.StatusNotFound, &resp)
			return
		}
		if destConn, err = dbutil.GetConn(ds.Vendor, ds.Location, false); err != nil {
			resp.Error = fmt.Sprintf("failed to connect to destination: %v", err)
			response.WriteJSON(w, http.StatusInternalServerError, &resp)
			return
		}
	}

	// Create row reader based on origin endpoint
	src, err := copydata.NewRowReader(req.Origin, originConn, originFile)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	// Create row writer based on destination endpoint
	var destWriter io.Writer
	if req.Destination.Type == "file" {
		destWriter = w // use http.ResponseWriter as destWriter to stream file to client
	}
	dst, err := copydata.NewRowWriter(req.Destination, destConn, dbVendor, destWriter, src.Fields())
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	// Prepare to stream file.
	if req.Destination.Type == "file" {
		ext := req.Destination.Format
		if len(ext) > 4 {
			ext = req.Destination.Format[:4]
		}
		w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+"."+ext)
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Copy data
	resp.Data.Reads, resp.Data.Writes, err = copydata.CopyData(src, dst)

	// End file response
	if req.Destination.Type == "file" {
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
