package internaldb

import (
	"database/sql"
	"db-portal/internal/meta"
	"db-portal/internal/types"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver" // sqlite3 driver for internal config storage
	"golang.org/x/crypto/bcrypt"
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

	return &Store{
		DBPath: dbPath,
		DB:     db,
	}, nil
}

func (s *Store) WarmUp() error {
	row := s.DB.QueryRow("SELECT 1")
	var i int
	return row.Scan(&i)
}

// ConnDetails holds the result of the query.
type ConnDetails struct {
	DBVendor types.DBVendor
	DSN      string
}

// FetchConn retrieves the dbtype and dsn for a given user and connection name.
func (s *Store) FetchConn(username, name string) (ConnDetails, error) {

	if name == "db-portal" {
		return ConnDetails{DBVendor: "sqlite3", DSN: s.DBPath}, nil
	}

	query := `
        select     c.dbtype, c.dsn
        from       user_connection uc
        inner join user u       on u.id = uc.user_id
        inner join connection c on c.id = uc.connection_id
        where      u.name = ? 
          and      c.name = ?`
	var details ConnDetails
	err := s.DB.QueryRow(query, username, name).Scan(&details.DBVendor, &details.DSN)
	return details, err
}

func (s *Store) FetchUserConns(username string) ([][]string, error) {
	query := `
        select     c.name, c.dbtype
        from       user_connection uc
        inner join user u       on u.id = uc.user_id
        inner join connection c on c.id = uc.connection_id
        where      u.name = ?
        order by   c.dbtype, c.name`

	rows, err := s.DB.Query(query, username)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return connections, nil
}

// checks if the provided username and password match the stored credentials.
func (s *Store) CheckUserCredentials(username, password string) (bool, error) {
	query := `select pwdhash from user where name = ?`
	var pwdhash string
	err := s.DB.QueryRow(query, username).Scan(&pwdhash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(pwdhash), []byte(password))
	return err == nil, nil
}
