package auth

import (
	"database/sql"
	"db-portal/internal/meta"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver" // sqlite3 driver for internal config storage
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserContextKey = contextKey("username")

var dbPath string
var db *sql.DB

func Initialize(configPath string) {
	var err error
	dbPath = filepath.Clean(filepath.Join(configPath, meta.AppName+".db"))
	_, err = os.Stat(dbPath)
	if err != nil {
		panic("failed to find database file : " + dbPath + ". Error :  " + err.Error())
	} else {
		fmt.Printf("DB file %v will be used\n", dbPath)
	}
	db, _ = sql.Open("sqlite3", dbPath)
}

// ConnectionDetails holds the result of the query.
type ConnectionDetails struct {
	DBType string
	DSN    string
}

// GetConnectionDetails retrieves the dbtype and dsn for a given user and connection name.
func GetConnectionDetails(username, name string) (details ConnectionDetails, err error) {

	if name == "db-portal" {
		details.DBType = "sqlite3"
		details.DSN = dbPath
		return
	}

	query := `
        select     c.dbtype, c.dsn
        from       user_connection uc
        inner join user u       on u.id = uc.user_id
        inner join connection c on c.id = uc.connection_id
        where      u.name = ? 
		  and      c.name = ?`
	err = db.QueryRow(query, username, name).Scan(&details.DBType, &details.DSN)

	return
}

func GetUserConnections(username string) ([][]string, error) {
	query := `
        select     c.name, c.dbtype
        from       user_connection uc
        inner join user u       on u.id = uc.user_id
        inner join connection c on c.id = uc.connection_id
        where      u.name = ?
		order by   c.dbtype, c.name`

	rows, err := db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections [][]string
	for rows.Next() {
		var name, dbtype string
		if err := rows.Scan(&name, &dbtype); err != nil {
			return nil, err
		}
		connections = append(connections, []string{name, dbtype})
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return connections, nil
}

// checks if the provided username and password match the stored credentials.
func checkCredentials(username, password string) (bool, error) {

	query := `select pwdhash from user where name = ?`
	// Execute the query
	var pwdhash string
	err := db.QueryRow(query, username).Scan(&pwdhash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(pwdhash), []byte(password))
	return err == nil, nil
}
