package copydata

import (
	"context"
	"database/sql"
	"db-portal/internal/dbutil"
	"errors"
	"io"
)

// dbRowReader implements RowReader for database sources.
type dbRowReader struct {
	rows   *sql.Rows
	fields []string
	types  []string
}

func NewDBRowReader(conn *sql.Conn, query string, args ...any) (RowReader, error) {
	if conn == nil {
		return nil, errors.New("db connection is nil")
	}

	rows, err := conn.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		rows.Close()
		return nil, err
	}

	types := make([]string, len(cols))
	if columnTypes, err := rows.ColumnTypes(); err != nil {
		for i, typ := range columnTypes {
			databaseTypeName := typ.DatabaseTypeName()
			types[i] = SQLToGenericType(databaseTypeName)
		}
	}

	return &dbRowReader{
		rows:   rows,
		fields: cols,
		types:  types,
	}, nil
}

func (r *dbRowReader) ReadRow() (Row, error) {
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	cols := make([]any, len(r.fields))
	colPtrs := make([]any, len(r.fields))
	for i := range cols {
		colPtrs[i] = &cols[i]
	}
	if err := r.rows.Scan(colPtrs...); err != nil {
		return nil, err
	}
	return cols, nil
}

func (r *dbRowReader) Fields() []string { return r.fields }
func (r *dbRowReader) Types() []string  { return r.types }
func (r *dbRowReader) Close() error     { return r.rows.Close() }

// dbRowWriter implements RowWriter for database destinations.
type dbRowWriter struct {
	conn    *sql.Conn
	stmt    *sql.Stmt
	table   string
	columns []string
}

func NewDBRowWriter(conn *sql.Conn, dbVendor string, table string, columns []string) (RowWriter, error) {
	if conn == nil {
		return nil, errors.New("db connection is nil")
	}
	if table == "" || len(columns) == 0 {
		return nil, errors.New("table name or columns missing")
	}

	// Build insert statement
	placeholders, err := dbutil.SetPlaceholders(dbVendor, len(columns))
	if err != nil {
		return nil, err
	}
	query := "INSERT INTO " + table + " (" + joinColumns(columns) + ") VALUES (" + joinColumns(placeholders) + ")"
	//fmt.Println("Prepared SQL Query:", query)

	stmt, err := conn.PrepareContext(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return &dbRowWriter{
		conn:    conn,
		stmt:    stmt,
		table:   table,
		columns: columns,
	}, nil
}

func (w *dbRowWriter) WriteFields(fields []string, types []string) error {
	//w.columns = fields
	return nil
}

func (w *dbRowWriter) WriteRow(row Row) error {
	_, err := w.stmt.ExecContext(context.Background(), row...)
	return err
}

func (w *dbRowWriter) Close() error {
	if w.stmt != nil {
		return w.stmt.Close()
	}
	return nil
}

// Helper to join columns for SQL
func joinColumns(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	out := cols[0]
	for _, c := range cols[1:] {
		out += "," + c
	}
	return out
}
