package copydata

import (
	"database/sql"
	"fmt"
	"io"
)

func NewRowReader(ep DataEndpoint, conn *sql.Conn, file io.Reader) (RowReader, error) {
	switch ep.Type {
	case "table":
		query := "select * from " + ep.Table
		return NewDBRowReader(conn, query)
	case "query":
		return NewDBRowReader(conn, ep.Query)
	case "file":
		switch ep.Format {
		case "csv":
			return NewCSVRowReader(file)
		case "json":
			return NewJSONRowReader(file)
		case "jsonTabular":
			return NewJSONTabularRowReader(file)
		case "xlsx":
			return NewXLSXRowReader(file)
		}
	}
	return nil, fmt.Errorf("unsupported reader. type: %s, format: %s", ep.Type, ep.Format)
}

func NewRowWriter(ep DataEndpoint, conn *sql.Conn, dbVendor string, file io.Writer, fields []string) (RowWriter, error) {
	switch ep.Type {
	case "table":
		return NewDBRowWriter(conn, dbVendor, ep.Table, fields)
	case "file":
		switch ep.Format {
		case "csv":
			return NewCSVRowWriter(file)
		case "json":
			return NewJSONRowWriter(file)
		case "jsonTabular":
			return NewJSONTabularRowWriter(file)
		case "xlsx":
			return NewXLSXRowWriter(file)
		}
	}
	return nil, fmt.Errorf("unsupported writer. type: %s, format: %s", ep.Type, ep.Format)
}
