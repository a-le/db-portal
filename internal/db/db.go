package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"       // mysql/mariaDB driver
	_ "github.com/jackc/pgx/v5/stdlib"       // PostgreSQL driver
	_ "github.com/microsoft/go-mssqldb"      // mssql driver
	_ "github.com/nakagami/firebirdsql"      // firebirdsql driver
	_ "github.com/ncruces/go-sqlite3/driver" // sqlite3 driver
	_ "github.com/ncruces/go-sqlite3/embed"  // sqlite3 driver
)

// Conn is a wrapper around *sql.Conn
type Conn struct {
	*sql.Conn
	DBType string
	Driver string
}

func DriverName(dbType string) (driverName string, err error) {
	Drivers := map[string]string{
		"firebird":   "firebirdsql",
		"mysql":      "mysql",
		"mssql":      "sqlserver",
		"postgresql": "pgx",
		"sqlite3":    "sqlite3",
	}
	driverName = Drivers[dbType]
	if driverName == "" {
		err = fmt.Errorf("unknown database type: %s", dbType)
	}
	return
}

// Cache for *sql.DB (DB is a database handle representing a pool of zero or more underlying connections.)
var dbCache = struct {
	sync.Mutex
	dbs map[string]*sql.DB
}{dbs: make(map[string]*sql.DB)}

// GetConn returns a connection from dbCache with useCache = true, else it returns a new connection.
func GetConn(DBType string, DSN string, useCache bool) (conn Conn, err error) {

	var driverName string
	if driverName, err = DriverName(DBType); err != nil {
		return
	}

	conn.DBType = DBType
	conn.Driver = driverName

	var db *sql.DB
	var cacheHit bool
	if useCache {
		dbCache.Lock()
		defer dbCache.Unlock()
		db, cacheHit = dbCache.dbs[fmt.Sprintf("%s:%s", DBType, DSN)]
	}

	// Open a new database connection pool
	if !cacheHit {
		//fmt.Printf("initializing a new connection pool for the %s database with the %s Go driver\n", DBType, driverName)
		db, err = sql.Open(driverName, DSN)
		if err != nil {
			return
		}
		if !useCache {
			defer db.Close()
		}
	}

	// Get a single connection from the pool
	conn.Conn, err = db.Conn(context.Background())
	if err != nil {
		if cacheHit {
			db.Close()
			delete(dbCache.dbs, fmt.Sprintf("%s:%s", DBType, DSN))
		}
		return
	}

	if useCache && !cacheHit {
		// put the database connection pool in cache
		dbCache.dbs[fmt.Sprintf("%s:%s", DBType, DSN)] = db
	}

	return
}

// QResult is a struct that holds the result of a QQuery.
type QResult struct {
	Rows         [][]any       `json:"rows"`
	Cols         []string      `json:"cols"`
	DatabaseType []string      `json:"databaseType"`
	GoScanType   []string      `json:"goScanType"`
	Duration     time.Duration `json:"duration"`
	RowsAffected int64         `json:"rowsAffected"`
	Truncated    bool          `json:"truncated"`
	DBerror      error         `json:"DBerror"`
}

// Custom JSON marshaling to handle DBerror field.
func (q QResult) MarshalJSON() ([]byte, error) {
	type Alias QResult // Create an alias to avoid infinite recursion
	return json.Marshal(&struct {
		DBerror string `json:"DBerror"` // Convert DBerror to a string
		*Alias
	}{
		DBerror: func() string {
			if q.DBerror != nil {
				return q.DBerror.Error()
			}
			return ""
		}(),
		Alias: (*Alias)(&q), // Embed the rest of the fields
	})
}

func NewQResult() QResult {
	return QResult{
		Rows:         [][]any{},
		Cols:         []string{},
		DatabaseType: []string{},
		GoScanType:   []string{},
		Duration:     0,
		RowsAffected: 0,
		Truncated:    false,
		DBerror:      nil,
	}
}

func StatementType(query string) (statementType string) {
	matches := regexp.MustCompile(`^\s*(\w+)`).FindStringSubmatch(strings.TrimSpace(query))
	if len(matches) > 1 {
		match := strings.ToLower(matches[1])
		if slices.Contains([]string{"select", "with", "explain", "analyze", "show", "describe", "analyze", "exec"}, match) {
			return "query"
		}
	}
	return "exec"
}

func Exec(conn Conn, query string, args []any) (result sql.Result, err error) {
	result, err = conn.ExecContext(context.Background(), query, args...)
	return
}

func Query(ctx context.Context, conn Conn, query string, args []any) (rows *sql.Rows, err error) {
	rows, err = conn.QueryContext(ctx, query, args...)
	if err != nil && rows != nil {
		rows.Close()
	}
	return
}

func QQuery(ctx context.Context, conn Conn, query string, args []any, limit int, statementType string) (qResult QResult, err error) {

	qResult = NewQResult()
	var rows *sql.Rows

	if statementType == "auto" {
		statementType = StatementType(query)
	}

	/*
		Exec statement
	*/
	if statementType == "exec" {
		startTime := time.Now()
		var result sql.Result
		result, qResult.DBerror = Exec(conn, query, args)
		qResult.Duration = time.Since(startTime)
		if qResult.DBerror != nil {
			return
		}

		qResult.RowsAffected, _ = result.RowsAffected()
		qResult.Rows = append(qResult.Rows, []any{qResult.RowsAffected})
		qResult.Cols = []string{"rows_affected"}
		qResult.DatabaseType = []string{"int64"}
		qResult.GoScanType = []string{"int64"}

		return
	}

	/*
		Query statement
	*/
	startTime := time.Now()
	rows, qResult.DBerror = Query(ctx, conn, query, args)
	qResult.Duration = time.Since(startTime)
	if qResult.DBerror != nil {
		return
	}
	defer rows.Close()

	// Retrieve column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return
	}
	// Populate qResult.Cols with column information
	for _, c := range columnTypes {
		qResult.Cols = append(qResult.Cols, c.Name())
		qResult.DatabaseType = append(qResult.DatabaseType, c.DatabaseTypeName())
		if c.ScanType() != nil {
			// nil will make .String() panic
			// test case: firebird sample base EMPLOYEE: SELECT LANGUAGE_REQ from "JOB" (LANGUAGE_REQ is of sql type is ARRAY)
			qResult.GoScanType = append(qResult.GoScanType, c.ScanType().String())
		} else {
			qResult.GoScanType = append(qResult.GoScanType, "unknown")
		}
	}

	// populate qResult.Rows
	for rows.Next() {
		if limit > 0 && qResult.RowsAffected >= int64(limit) {
			qResult.Truncated = true
			return // Limit the number of rows returned
		}

		values := make([]any, len(qResult.Cols))
		valuePtrs := make([]any, len(qResult.Cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err = rows.Scan(valuePtrs...); err != nil {
			return
		}

		row := make([]any, len(qResult.Cols))
		for i := range qResult.Cols {
			row[i] = values[i]

			if conn.DBType == "mysql" {
				byteArray, ok := values[i].([]byte)
				if ok {
					row[i] = string(byteArray)
				}
			}

		}
		qResult.Rows = append(qResult.Rows, row)
		qResult.RowsAffected++
	}

	return
}
