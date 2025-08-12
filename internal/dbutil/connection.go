package dbutil

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// Cache for *sql.DB (DB is a database handle representing a pool of zero or more underlying connections.)
var dbCache = struct {
	sync.Mutex
	dbs map[string]*sql.DB
}{dbs: make(map[string]*sql.DB)}

// GetConn returns a connection from dbCache with useCache = true, else it returns a new connection.
func GetConn(vendor string, location string, useCache bool) (conn *sql.Conn, err error) {

	var driverName string
	if driverName, err = DriverName(vendor); err != nil {
		return
	}

	var db *sql.DB
	var cacheHit bool
	if useCache {
		dbCache.Lock()
		defer dbCache.Unlock()
		db, cacheHit = dbCache.dbs[fmt.Sprintf("%s:%s", vendor, location)]
	}

	// Open a new database connection pool
	if !cacheHit {
		db, err = sql.Open(driverName, location)
		if err != nil {
			return
		}
		if !useCache {
			defer db.Close()
		}
	}

	// Get a single connection from the pool
	conn, err = db.Conn(context.Background())
	if err != nil {
		if cacheHit {
			db.Close()
			delete(dbCache.dbs, fmt.Sprintf("%s:%s", vendor, location))
		}
		return
	}

	if useCache && !cacheHit {
		// put the database connection pool in cache
		dbCache.dbs[fmt.Sprintf("%s:%s", vendor, location)] = db
	}

	return
}
