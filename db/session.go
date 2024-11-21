package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"morbo/errors"
	"morbo/log"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (db *DB) AuthenticateByCredentials(credentials Credentials) (userID int, statusCode int, err error) {
	var hashedPassword string

	query := `SELECT id, password FROM users WHERE username = ?`
	row := db.pool.QueryRow(context.Background(), query, credentials.Username)
	err = row.Scan(&userID, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info.Println("no such user found")
			return -1, http.StatusUnauthorized, errors.Error
		}
		log.Error.Println(err)
		log.Error.Println("failed to authenticate via credentials")
		return -1, http.StatusInternalServerError, errors.Error
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password))
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("password doesn't match")
		return -1, http.StatusUnauthorized, errors.Error
	}

	return userID, 0, nil
}

func (db *DB) AuthenticateBySessionToken(sessionToken string) (userID int, statusCode int, err error) {
	query := `SELECT user_id FROM sessions WHERE session_token = ?`
	row := db.pool.QueryRow(context.Background(), query, sessionToken)
	err = row.Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Error.Println("no such session token found")
			return -1, http.StatusUnauthorized, errors.Error
		}
		log.Error.Println(err)
		log.Error.Println("failed to authenticate via session token")
		return -1, http.StatusInternalServerError, errors.Error
	}

	return userID, 0, nil
}

func (db *DB) GenerateSessionToken(userID int) (sessionToken string, err error) {
	byteSessionToken := make([]byte, 40)
	if _, err := rand.Read(byteSessionToken); err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to generate a session token")
		return "", errors.Error
	}
	sessionToken = base64.RawURLEncoding.EncodeToString(byteSessionToken)

	query := `INSERT INTO sessions (session_token, user_id) VALUES (?, ?)`
	_, err = db.pool.Exec(context.Background(), query, sessionToken, userID)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to store a session token")
		return "", errors.Error
	}

	return sessionToken, nil
}

func (db *DB) DeleteSessionToken(sessionToken string) error {
	query := `DELETE FROM sessions WHERE session_token = ?`
	_, err := db.pool.Exec(context.Background(), query, sessionToken)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to delete a session token")
		return errors.Error
	}

	return nil
}
