package internaldb

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

func (s *Store) CheckIsAdmin(username string) (bool, error) {
	var isadmin int
	err := s.DB.QueryRow(`
        SELECT isadmin
        FROM user
        WHERE name = ?
    `, username).Scan(&isadmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // user not found, not admin
		}
		return false, err
	}
	return isadmin == 1, nil
}

type User struct {
	Name    string `json:"name"`
	IsAdmin int    `json:"isadmin"`
}

func (s *Store) GetAllUsers(currentUsername string) ([]User, error) {
	query := `
    WITH currentuser AS (
        SELECT name, isadmin 
        FROM user 
        WHERE name = ?
    ) 
    SELECT user.name, user.isadmin 
    FROM user 
    INNER JOIN currentuser ON user.name = currentuser.name OR currentuser.isadmin = 1
    ORDER BY 1`

	rows, err := s.DB.Query(query, currentUsername)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var isadmin int
		if err := rows.Scan(&user.Name, &isadmin); err != nil {
			return nil, err
		}
		//user.IsAdmin = isadmin == 1
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) CreateUser(currentUsername, username string, isAdmin int, password string) error {
	query := `
	WITH currentuser AS (
        SELECT name, isadmin 
        FROM user 
        WHERE name = ?
    )
	INSERT INTO user (name, isadmin, pwdhash)
	SELECT ?, ?, ? 
	FROM currentuser 
	WHERE isadmin = 1
	`
	// Hash the password before storing
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(query, currentUsername, username, isAdmin, string(pwdHash))
	return err
}

// checks if the provided username and password match the stored credentials.
// also returns UserInfo if found.
func (s *Store) CheckUserCredentials(username string, password string) (bool, *User, error) {
	query := `SELECT name, isadmin, pwdhash FROM user WHERE name = ?`
	var name string
	var isadmin int
	var pwdhash string
	if err := s.DB.QueryRow(query, username).Scan(&name, &isadmin, &pwdhash); err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, err
	}
	err := bcrypt.CompareHashAndPassword([]byte(pwdhash), []byte(password))
	if err != nil {
		return false, nil, nil
	}
	userInfo := &User{
		Name:    name,
		IsAdmin: isadmin,
	}
	return true, userInfo, nil
}

// SetMasterUserPassword updates the password hash for the master user (id=1).
func (s *Store) SetMasterUserPassword(newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`UPDATE user SET pwdhash = ? WHERE id = 1`, string(hash))
	return err
}
