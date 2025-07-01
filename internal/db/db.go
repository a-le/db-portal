package db

import (
	"context"
	"database/sql"
	"db-portal/internal/types"

	"fmt"
	"sync"

	_ "github.com/ClickHouse/clickhouse-go/v2" // clickhouse driver
	_ "github.com/go-sql-driver/mysql"         // mysql/mariaDB driver
	_ "github.com/jackc/pgx/v5/stdlib"         // PostgreSQL driver
	_ "github.com/microsoft/go-mssqldb"        // mssql driver

	//_ "github.com/nakagami/firebirdsql"        // firebirdsql driver
	_ "github.com/ncruces/go-sqlite3/driver" // sqlite3 driver
	_ "github.com/ncruces/go-sqlite3/embed"  // sqlite3 driver
)

// Conn is a wrapper around *sql.Conn
type Conn struct {
	*sql.Conn
	DBVendor types.DBVendor
	Driver   string
}

func DriverName(dbVendor types.DBVendor) (driverName string, err error) {
	Drivers := map[types.DBVendor]string{
		types.DBVendorClickHouse: "clickhouse",
		//types.DBVendorFirebird:   "firebirdsql", // less supported and tested
		types.DBVendorMySQL:    "mysql",
		types.DBVendorMariaDB:  "mysql",
		types.DBVendorMSSQL:    "sqlserver",
		types.DBVendorPostgres: "pgx",
		types.DBVendorSQLite:   "sqlite3",
	}
	driverName = Drivers[dbVendor]
	if driverName == "" {
		err = fmt.Errorf("unknown database type: %s", dbVendor)
	}
	return
}

// Cache for *sql.DB (DB is a database handle representing a pool of zero or more underlying connections.)
var dbCache = struct {
	sync.Mutex
	dbs map[string]*sql.DB
}{dbs: make(map[string]*sql.DB)}

// GetConn returns a connection from dbCache with useCache = true, else it returns a new connection.
func GetConn(dbVendor types.DBVendor, DSN string, useCache bool) (conn Conn, err error) {

	var driverName string
	if driverName, err = DriverName(dbVendor); err != nil {
		return
	}

	conn.DBVendor = dbVendor
	conn.Driver = driverName

	var db *sql.DB
	var cacheHit bool
	if useCache {
		dbCache.Lock()
		defer dbCache.Unlock()
		db, cacheHit = dbCache.dbs[fmt.Sprintf("%s:%s", dbVendor, DSN)]
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
			delete(dbCache.dbs, fmt.Sprintf("%s:%s", dbVendor, DSN))
		}
		return
	}

	if useCache && !cacheHit {
		// put the database connection pool in cache
		dbCache.dbs[fmt.Sprintf("%s:%s", dbVendor, DSN)] = db
	}

	return
}

func ExecContext(conn Conn, query string, args []any) (result sql.Result, err error) {
	result, err = conn.ExecContext(context.Background(), query, args...)
	return
}

func QueryContext(ctx context.Context, conn Conn, query string, args []any) (rows *sql.Rows, err error) {
	rows, err = conn.QueryContext(ctx, query, args...)
	if err != nil && rows != nil {
		rows.Close()
	}
	return
}
