package internaldb

import (
	"database/sql"
	"db-portal/internal/dbutil"
)

type DataSource struct {
	Name     string `json:"name"`
	Vendor   string `json:"vendor"`
	Location string `json:"location"`
}

func (s *Store) CreateDataSource(currentUsername, name, vendor, location string) error {
	query := `
	WITH currentuser AS (
        SELECT name, isadmin 
        FROM user 
        WHERE name = ?
    )
	INSERT INTO ds (name, vendor_id, location)
	SELECT ?, vendor.id, ? 
	FROM currentuser 
	INNER JOIN vendor ON vendor.name = ?
	WHERE isadmin = 1
	`

	_, err := s.DB.Exec(query, currentUsername, name, location, vendor)
	return err
}

func (s *Store) TestDataSource(vendor, location string) (bool, error) {
	driverName, err := dbutil.DriverName(vendor)
	if err != nil {
		return false, err
	}
	conn, err := sql.Open(driverName, location)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// Try to ping the database to check if the connection is valid
	if err := conn.Ping(); err != nil {
		return false, err
	}
	return true, nil
}
