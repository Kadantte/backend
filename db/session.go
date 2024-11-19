package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	Unathorized         = errors.New("unauthorized")
	InternalServerError = errors.New("internal server error")
)

type UserID = int

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (db *DB) AuthenticateByCredentials(credentials Credentials) (UserID, error) {
	var userID UserID
	var hashedPassword string

	query := `SELECT id, password FROM users WHERE username = ?`
	row := db.pool.QueryRow(context.Background(), query, credentials.Username)
	err := row.Scan(&userID, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, errors.Join(Unathorized, err)
		}
		return -1, errors.Join(InternalServerError, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password))
	if err != nil {
		return -1, errors.Join(Unathorized, err)
	}

	return userID, nil
}

func (db *DB) AuthenticateBySessionToken(sessionToken string) (UserID, error) {
	var userID UserID

	query := `SELECT user_id FROM sessions WHERE session_token = ?`
	row := db.pool.QueryRow(context.Background(), query, sessionToken)
	err := row.Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, errors.Join(Unathorized, err)
		}
		return -1, errors.Join(InternalServerError, err)
	}

	return userID, nil
}

func (db *DB) CreateSessionToken(userID UserID) (string, error) {
	byteSessionToken := make([]byte, 40)
	if _, err := rand.Read(byteSessionToken); err != nil {
		return "", errors.Join(fmt.Errorf("failed to create a session token"), err)
	}
	sessionToken := base64.RawURLEncoding.EncodeToString(byteSessionToken)

	query := `INSERT INTO sessions (session_token, user_id) VALUES (?, ?)`
	_, err := db.pool.Exec(context.Background(), query, sessionToken, userID)
	if err != nil {
		return "", errors.Join(fmt.Errorf("failed to store a session token"), err)
	}

	return sessionToken, nil
}

func (db *DB) DeleteSessionToken(sessionToken string) error {
	query := `DELETE FROM sessions WHERE session_token = ?`
	_, err := db.pool.Exec(context.Background(), query, sessionToken)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to delete a session token"), err)
	}

	return nil
}
