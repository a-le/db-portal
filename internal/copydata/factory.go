package copydata

import (
	"context"
	"database/sql"
	"fmt"
	"io"
)

func NewRowReader(ep EndPoint, ctx context.Context, conn *sql.Conn, file io.Reader) (RowReader, error) {
	switch ep.Type {
	case "table":
		query := "select * from " + ep.Table
		return NewDBRowReader(ctx, conn, ep.DBVendor, query)
	case "query":
		return NewDBRowReader(ctx, conn, ep.DBVendor, ep.Query)
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

func NewRowWriter(ep EndPoint, ctx context.Context, tx *sql.Tx, file io.Writer, fields []string) (RowWriter, error) {
	switch ep.Type {
	case "table":
		createTable := (ep.IsNewTable == "1")
		// transaction is managed by the caller, not the writer
		return NewDBRowWriter(ctx, tx, ep.DBVendor, ep.Table, createTable, fields)
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
