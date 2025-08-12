/*
Access Control Rules:
The current user, unless he is admin (user.isadmin = 1), cannot retrieve or modify:
- data sources (ds.*) not registred to himself
- users (user.*) but himself
- ds.location, unless vendor.name = 'sqlite3'
- user data source (user_ds.*)
*/
package internaldb

import (
	"database/sql"
	"db-portal/internal/meta"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver" // sqlite3 driver for internal config storage
)

type Store struct {
	DBPath string
	DB     *sql.DB
}

func NewStore(configPath string) (*Store, error) {

	dbPath := filepath.Clean(filepath.Join(configPath, meta.AppName+".db"))

	if _, err := os.Stat(dbPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// internal DB always exists as data source
	// update data source location of internaldb (ds.id=1) to its current path
	_, err = db.Exec("UPDATE ds SET location = ? WHERE id = 1", dbPath)
	if err != nil {
		return nil, err
	}

	return &Store{
		DBPath: dbPath,
		DB:     db,
	}, nil
}
