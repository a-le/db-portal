package copydata

// todo: context.Background()...

import (
	"context"
	"database/sql"
	"db-portal/internal/dbutil"
	"errors"
	"io"
)

// dbRowReader implements RowReader for database sources.
type dbRowReader struct {
	rows    *sql.Rows
	columns []string
	types   []string
}

func NewDBRowReader(conn *sql.Conn, dbVendor string, query string, args ...any) (RowReader, error) {
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

	// Attempt to get each column's DatabaseTypeName.
	// If successful, infer the canonical type for each column.
	// Some drivers may return an empty string for DatabaseTypeName;
	// in such cases, the canonical type will default to "text" as a fallback for unknown types.
	types := make([]string, len(cols))
	if columnTypes, err := rows.ColumnTypes(); err == nil {
		for i, c := range columnTypes {
			databaseTypeName := c.DatabaseTypeName()
			types[i] = dbutil.CanonicalType(dbVendor, databaseTypeName)
		}
	}

	return &dbRowReader{
		rows:    rows,
		columns: cols,
		types:   types,
	}, nil
}

func (r *dbRowReader) ReadRow() (Row, error) {
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return nil, err
		}
		r.rows.Close()
		return nil, io.EOF
	}
	cols := make([]any, len(r.columns))
	colPtrs := make([]any, len(r.columns))
	for i := range cols {
		colPtrs[i] = &cols[i]
	}
	if err := r.rows.Scan(colPtrs...); err != nil {
		return nil, err
	}
	return cols, nil
}

func (r *dbRowReader) Fields() []string { return r.columns }
func (r *dbRowReader) Types() []string  { return r.types }

// dbRowWriter
type dbRowWriter struct {
	conn        *sql.Conn
	table       string
	createTable bool
	columns     []string
	types       []string
	batch       [][]any
	batchSize   int
	dbVendor    string
}

func NewDBRowWriter(conn *sql.Conn, dbVendor string, table string, createTable bool, columns []string) (RowWriter, error) {
	if conn == nil {
		return nil, errors.New("db connection is nil")
	}
	if table == "" || len(columns) == 0 {
		return nil, errors.New("table name or columns missing")
	}
	return &dbRowWriter{
		conn:        conn,
		table:       table,
		createTable: createTable,
		columns:     columns,
		batch:       make([][]any, 0, 100),
		batchSize:   100,
		dbVendor:    dbVendor,
	}, nil
}

func (w *dbRowWriter) WriteFields(columns []string, types []string) error {
	if !w.createTable {
		return nil
	}
	if len(columns) == 0 || len(types) != len(columns) {
		return errors.New("columns and types mismatch or empty")
	}

	// Build CREATE TABLE statement
	var colsDef []string
	for i, col := range columns {
		sqlType := dbutil.VendorType(w.dbVendor, types[i])
		colsDef = append(colsDef, col+" "+sqlType)
	}
	query := "CREATE TABLE " + w.table + " (" + joinColumns(colsDef) + ")"

	_, err := w.conn.ExecContext(context.Background(), query)
	return err
}

func (w *dbRowWriter) WriteRow(row Row) (rowsWritten int, err error) {
	w.batch = append(w.batch, row)
	if len(w.batch) >= w.batchSize {
		return w.Flush()
	}
	return
}

func (w *dbRowWriter) Flush() (rowsWritten int, err error) {
	if len(w.batch) == 0 {
		return
	}
	numRows := len(w.batch)
	numCols := len(w.columns)
	placeholders, err := dbutil.SetBatchPlaceholders(w.dbVendor, numCols, numRows)
	if err != nil {
		return
	}
	// Build VALUES clause
	valuesClause := ""
	for i := range numRows {
		if i > 0 {
			valuesClause += ","
		}
		start := i * numCols
		end := start + numCols
		valuesClause += "(" + joinColumns(placeholders[start:end]) + ")"
	}
	query := "INSERT INTO " + w.table + " (" + joinColumns(w.columns) + ") VALUES " + valuesClause
	args := []any{}
	for _, row := range w.batch {
		args = append(args, row...)
	}
	_, err = w.conn.ExecContext(context.Background(), query, args...)
	w.batch = w.batch[:0]
	return numRows, err
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
