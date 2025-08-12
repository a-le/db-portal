package internaldb

import "fmt"

const dsBaseQuery = `
    WITH currentuser AS (
        SELECT name, isadmin
        FROM user
        WHERE name = ?
    )
    SELECT
        ds.name,
        vendor.name AS vendor,
        CASE
            WHEN currentuser.isadmin = 1 OR vendor.name = 'sqlite3'
                THEN ds.location
			ELSE ''
        END AS location
    FROM  user 
    INNER JOIN user_ds ON user_ds.user_id = user.id
    INNER JOIN ds ON ds.id = user_ds.ds_id
    INNER JOIN vendor ON vendor.id = ds.vendor_id
    INNER JOIN currentuser ON user.name = currentuser.name OR currentuser.isadmin = 1
    WHERE 
`

func (s *Store) RequireUserDataSource(currentUsername, username, dsName string) (DataSource, error) {
	ds, err := s.GetUserDataSource(currentUsername, username, dsName)
	if err != nil {
		return DataSource{}, err
	}
	if ds == (DataSource{}) {
		return DataSource{}, fmt.Errorf("data source %q not found or not allowed for user %q", dsName, username)
	}
	return ds, nil
}

// Get user data source by its name.
func (s *Store) GetUserDataSource(currentUsername, username, dsName string) (DataSource, error) {
	query := dsBaseQuery + ` user.name = ? AND ds.name = ?`
	var result DataSource
	err := s.DB.QueryRow(query, currentUsername, username, dsName).Scan(&result.Name, &result.Vendor, &result.Location)
	return result, err
}

// Fetch DS registred to a user.
func (s *Store) GetAllUserDataSources(currentUsername, username string) ([]DataSource, error) {
	query := dsBaseQuery + `  user.name = ? ORDER BY 2, 1`

	rows, err := s.DB.Query(query, currentUsername, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DataSource
	for rows.Next() {
		var ds DataSource
		if err := rows.Scan(&ds.Name, &ds.Vendor, &ds.Location); err != nil {
			return nil, err
		}
		result = append(result, ds)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// Fetch DS not registred to a user
func (s *Store) GetAllUserAvailableDataSources(currentUsername, username string) ([]DataSource, error) {
	query := `
	WITH currentuser AS (
		SELECT name, isadmin
		FROM user
		WHERE name = ?
	)
	SELECT 
		ds.name,
		vendor.name AS vendor,
		ds.location
	FROM ds
	INNER JOIN vendor ON vendor.id = ds.vendor_id
	INNER JOIN currentuser ON currentuser.isadmin = 1
	WHERE ds.id NOT IN (
		SELECT user_ds.ds_id
		FROM user_ds
		INNER JOIN user ON user.id = user_ds.user_id
		WHERE user.name = ?
	)
	`
	rows, err := s.DB.Query(query, currentUsername, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DataSource
	for rows.Next() {
		var ds DataSource
		if err := rows.Scan(&ds.Name, &ds.Vendor, &ds.Location); err != nil {
			return nil, err
		}
		result = append(result, ds)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) CreateUserDataSource(currentUsername, username, dsName string) error {
	query := `
	WITH currentuser AS (
        SELECT name, isadmin 
        FROM user 
        WHERE name = ?
    )
	INSERT INTO user_ds (user_id, ds_id)
	SELECT user.id, ds.id
	FROM user
	INNER JOIN ds ON ds.name = ?
	INNER JOIN currentuser ON currentuser.isadmin = 1
	WHERE user.name = ?
	`

	_, err := s.DB.Exec(query, currentUsername, dsName, username)
	return err
}

func (s *Store) DeleteUserDataSource(currentUsername, username, dsName string) error {
	query := `
	WITH currentuser AS (
        SELECT name, isadmin 
        FROM user 
        WHERE name = ?
    )
	DELETE FROM user_ds
	WHERE (SELECT isadmin FROM currentuser) = 1
	AND ds_id = (SELECT id FROM ds WHERE name = ?)
	AND user_id = (SELECT id FROM user WHERE name = ?)
	`
	_, err := s.DB.Exec(query, currentUsername, dsName, username)
	return err
}
