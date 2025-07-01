package datatransfer

import (
	"compress/gzip"
	"database/sql"
	"db-portal/internal/db"
	"db-portal/internal/types"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Exporter interface {
	ExportCSV(w http.ResponseWriter, rows *sql.Rows, gz bool) error
	ExportXLSX(w http.ResponseWriter, rows *sql.Rows, gz bool) error
	ExportJSON(w http.ResponseWriter, rows *sql.Rows, gz bool) error
	ExportJSONTabular(w http.ResponseWriter, rows *sql.Rows, gz bool) error
	SetDBVendor(v types.DBVendor)
}

type DefaultExporter struct {
	DBVendor types.DBVendor
}

func (e *DefaultExporter) SetDBVendor(v types.DBVendor) {
	e.DBVendor = v
}

func (e *DefaultExporter) exportFile(
	w http.ResponseWriter,
	rows *sql.Rows,
	ext string,
	contentType string,
	exportFunc func(*sql.Rows, *os.File, types.DBVendor) error,
	gzipEnabled bool,
) error {

	tmpfile, err := os.CreateTemp("", "dbexport_*."+ext)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if err := exportFunc(rows, tmpfile, e.DBVendor); err != nil {
		return err
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		return err
	}

	if gzipEnabled {
		// Create a new temp file for the compressed file
		gztmpfile, err := os.CreateTemp("", "dbexport_*."+ext+".gz")
		if err != nil {
			return err
		}
		defer os.Remove(gztmpfile.Name())
		defer gztmpfile.Close()

		// Compress tmpfile into gzfile
		gzWriter := gzip.NewWriter(gztmpfile)
		if _, err := io.Copy(gzWriter, tmpfile); err != nil {
			gzWriter.Close()
			return err
		}
		gzWriter.Close()

		if _, err := gztmpfile.Seek(0, 0); err != nil {
			return err
		}

		info, _ := gztmpfile.Stat()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+"."+ext+".gz")
		_, err = io.Copy(w, gztmpfile)
		return err
	}

	info, _ := tmpfile.Stat()
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+"."+ext)
	_, err = io.Copy(w, tmpfile)
	return err
}

func (e *DefaultExporter) ExportCSV(w http.ResponseWriter, rows *sql.Rows, gz bool) error {
	exportFunc := func(rows *sql.Rows, file *os.File, _ types.DBVendor) error {
		return db.RowsToCsv(rows, file)
	}
	return e.exportFile(w, rows, "csv", "text/csv", exportFunc, gz)
}

func (e *DefaultExporter) ExportJSON(w http.ResponseWriter, rows *sql.Rows, gz bool) error {
	exportFunc := func(rows *sql.Rows, file *os.File, _ types.DBVendor) error {
		return db.RowsToJson(rows, file)
	}
	return e.exportFile(w, rows, "json", "application/json", exportFunc, gz)
}

func (e *DefaultExporter) ExportJSONTabular(w http.ResponseWriter, rows *sql.Rows, gz bool) error {
	exportFunc := func(rows *sql.Rows, file *os.File, dbVendor types.DBVendor) error {
		return db.RowsToJsonTabular(rows, file, dbVendor)
	}
	return e.exportFile(w, rows, "json", "application/json", exportFunc, gz)
}

func (e *DefaultExporter) ExportXLSX(w http.ResponseWriter, rows *sql.Rows, gz bool) error {
	exportFunc := func(rows *sql.Rows, file *os.File, _ types.DBVendor) error {
		return db.RowsToXlsx(rows, file.Name())
	}
	return e.exportFile(w, rows, "xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", exportFunc, gz)
}
