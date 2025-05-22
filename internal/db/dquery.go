package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type DResult struct {
	Cols          []string      `json:"cols"`
	DatabaseTypes []string      `json:"databaseTypes"`
	DBerror       error         `json:"DBerror"`
	Duration      time.Duration `json:"duration"`
	Rows          [][]any       `json:"rows"`
	RowsAffected  int64         `json:"rowsAffected"` // Number of rows affected by exec
	RowsReturned  int64         `json:"rowsReturned"` // Number of rows returned by query
	StmtCmd       string        `json:"stmtCmd"`      // Top level SQL keyword : update|insert|delete|select|create|drop|alter|show...
	StmtType      string        `json:"stmtType"`     // Identified type of statement : "query"|"not-query"
	Truncated     bool          `json:"truncated"`
}

func (q DResult) MarshalJSON() ([]byte, error) {
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
		DBerror: func() string {
			if q.DBerror != nil {
				return q.DBerror.Error()
			}
			return ""
		}(),
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

func DExecContext(ctx context.Context, conn Conn, query string, args []any) (dResult DResult, err error) {

	startTime := time.Now()
	var result sql.Result
	result, dResult.DBerror = conn.ExecContext(context.Background(), query, args...)
	dResult.Duration = time.Since(startTime)
	if dResult.DBerror != nil {
		return
	}
	dResult.RowsAffected, _ = result.RowsAffected()
	return
}

func DQueryContext(ctx context.Context, conn Conn, query string, args []any, limit int64) (dResult DResult, err error) {

	var rows *sql.Rows
	startTime := time.Now()
	rows, dResult.DBerror = conn.QueryContext(ctx, query, args...)
	dResult.Duration = time.Since(startTime)

	if rows != nil {
		defer rows.Close()
	}

	if dResult.DBerror != nil {
		return
	}

	// Retrieve column names
	if dResult.Cols, err = rows.Columns(); err != nil {
		return
	}

	// process rows
	i := int64(0)
	for rows.Next() {
		i++
		if limit > 0 && i >= limit {
			dResult.Truncated = true
			continue // Limit the number of rows returned
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
			row[i] = values[i]

			if conn.DBType == "mysql" {
				if byteArray, ok := values[i].([]byte); ok {
					row[i] = string(byteArray)
				}
			}
		}
		dResult.Rows = append(dResult.Rows, row)
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

	// Populate qResult.DatabaseType and qResult.GoScanType
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
