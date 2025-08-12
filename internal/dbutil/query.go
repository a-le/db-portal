package dbutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2" // clickhouse driver
	_ "github.com/go-sql-driver/mysql"         // mysql/mariaDB driver
	_ "github.com/jackc/pgx/v5/stdlib"         // PostgreSQL driver
	_ "github.com/microsoft/go-mssqldb"        // mssql driver
	_ "github.com/ncruces/go-sqlite3/driver"   // sqlite3 driver
	_ "github.com/ncruces/go-sqlite3/embed"    // sqlite3 driver
)

type DBResult struct {
	Cols          []string      `json:"cols"`
	DatabaseTypes []string      `json:"databaseTypes"`
	DBerror       string        `json:"DBerror"`
	Duration      time.Duration `json:"duration"`
	Rows          [][]any       `json:"rows"`
	RowsAffected  int64         `json:"rowsAffected"` // Number of rows affected by exec
	RowsReturned  int64         `json:"rowsReturned"` // Number of rows returned by query
	StmtCmd       string        `json:"stmtCmd"`      // Top level SQL keyword : update|insert|delete|select|create|drop|alter|show...
	StmtType      string        `json:"stmtType"`     // Identified type of statement : "query"|"not-query"
	Truncated     bool          `json:"truncated"`
}

func (q *DBResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Cols          []string      `json:"cols"`
		DatabaseTypes []string      `json:"databaseTypes"`
		DBerror       string        `json:"DBerror"`
		Duration      time.Duration `json:"duration"`
		Rows          [][]any       `json:"rows"`
		RowsAffected  int64         `json:"rowsAffected"`
		RowsReturned  int64         `json:"rowsReturned"`
		StmtCmd       string        `json:"stmtCmd"`
		StmtType      string        `json:"stmtType"`
		Truncated     bool          `json:"truncated"`
	}{
		Cols: func() []string {
			if q.Cols == nil {
				return make([]string, 0)
			}
			return q.Cols
		}(),
		DatabaseTypes: func() []string {
			if q.DatabaseTypes == nil {
				return make([]string, 0)
			}
			return q.DatabaseTypes
		}(),
		DBerror:  q.DBerror,
		Duration: q.Duration,
		Rows: func() [][]any {
			if q.Rows == nil {
				return make([][]any, 0)
			}
			return q.Rows
		}(),
		RowsAffected: q.RowsAffected,
		RowsReturned: q.RowsReturned,
		StmtCmd:      q.StmtCmd,
		StmtType:     q.StmtType,
		Truncated:    q.Truncated,
	})
}

func ExecWithResult(ctx context.Context, conn *sql.Conn, query string, args []any) (dResult DBResult, err error) {

	startTime := time.Now()
	var result sql.Result
	result, err = conn.ExecContext(context.Background(), query, args...)
	dResult.Duration = time.Since(startTime)
	if err != nil {
		dResult.DBerror = err.Error()
		return
	}
	dResult.RowsAffected, _ = result.RowsAffected()
	return
}

func QueryWithResult(ctx context.Context, conn *sql.Conn, query string, args []any, limit int64) (dResult DBResult, err error) {

	var rows *sql.Rows
	startTime := time.Now()
	rows, err = conn.QueryContext(ctx, query, args...)
	dResult.Duration = time.Since(startTime)

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		dResult.DBerror = err.Error()
		return
	}

	// Retrieve column names
	if dResult.Cols, err = rows.Columns(); err != nil {
		return
	}

	// process rows
	i := int64(0)
	for rows.Next() {
		if limit > 0 && i >= limit {
			dResult.Truncated = true
			break // Limit the number of rows returned
		}

		values := make([]any, len(dResult.Cols))
		valuePtrs := make([]any, len(dResult.Cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err = rows.Scan(valuePtrs...); err != nil {
			break
		}

		row := make([]any, len(dResult.Cols))
		for i := range dResult.Cols {
			switch v := values[i].(type) {
			case []byte:
				// Convert byte slices to string for JSON compatibility
				row[i] = string(v)
			default:
				row[i] = values[i]
			}
		}
		dResult.Rows = append(dResult.Rows, row)

		i++
	}
	dResult.RowsReturned = i

	// Attempt to retrieve column types
	var columnTypes []*sql.ColumnType
	func() {
		defer func() {
			recover()
		}()
		// Call rows.ColumnTypes() inside the defer-protected block
		columnTypes, _ = rows.ColumnTypes()
	}()

	// Populate qResult.DatabaseType
	for i := range dResult.Cols {
		if i < len(columnTypes) {
			c := columnTypes[i]
			dResult.DatabaseTypes = append(dResult.DatabaseTypes, c.DatabaseTypeName())
		} else {
			// Default values if ColumnTypes() failed
			dResult.DatabaseTypes = append(dResult.DatabaseTypes, "unknown")
		}
	}

	return
}
