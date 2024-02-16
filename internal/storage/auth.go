package storage

import (
	"database/sql"
	"fmt"
)

func (s *Storage) Register(login string, pass string) (int64, error) {
	query := `
		SELECT id
		FROM USERS 
		WHERE LOGIN = $1
	`
	var userId int64
	if err := s.db.QueryRow(query, login).Scan(&userId); err != sql.ErrNoRows {
		if err != nil {
			return -1, fmt.Errorf("register user error -> %s", err)
		}
	}

	if userId != 0 {
		return -1, fmt.Errorf("login alrady exit")
	}

	query = `
		INSERT INTO USERS (login, pass) 
		VALUES($1, $2)
		RETURNING id
	`

	err := s.db.QueryRow(query, login, pass).Scan(&userId)
	if err != nil {
		return -1, fmt.Errorf("register user error -> %s", err)
	}

	return userId, err
}

func (s *Storage) Auth(login string, pass string) (int64, error) {
	query := `
	SELECT id
	FROM USERS 
	WHERE LOGIN = $1 AND PASS = $2
`
	var userId int64
	if err := s.db.QueryRow(query, login, pass).Scan(&userId); err != nil {
		if err == sql.ErrNoRows {
			return -1, fmt.Errorf("user not found")
		}
		return -1, fmt.Errorf("register user error -> %s", err)
	}
	return userId, nil
}
