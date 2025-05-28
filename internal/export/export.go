package export

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"time"

	"db-portal/internal/db"

	"github.com/joho/sqltocsv"
)

type Exporter interface {
	ExportCSV(w http.ResponseWriter, rows *sql.Rows) error
	ExportXLSX(w http.ResponseWriter, rows *sql.Rows) error
	ExportJSON(w http.ResponseWriter, rows *sql.Rows) error
}

type DefaultExporter struct{}

func (e *DefaultExporter) ExportCSV(w http.ResponseWriter, rows *sql.Rows) error {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+".csv")
	return sqltocsv.Write(w, rows)
}

// Export as array of objects in JSON format
func (e *DefaultExporter) ExportJSON(w http.ResponseWriter, rows *sql.Rows) error {
	jsonData, err := db.RowsToJson(rows)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+".json")

	_, err = w.Write(jsonData)
	return err
}

func (e *DefaultExporter) ExportXLSX(w http.ResponseWriter, rows *sql.Rows) error {
	tmpfile, err := os.CreateTemp("", "dbexport_*.xlsx")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if err := db.RowsToXlsx(rows, tmpfile.Name()); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=export_"+time.Now().Format("20060102-150405")+".xlsx")

	_, err = io.Copy(w, tmpfile)
	return err
}
