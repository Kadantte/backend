package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"morbo/context"
	"morbo/db"
	"morbo/errors"
	"morbo/log"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (conn *Connection) AuthenticateViaCredentials(credentials Credentials) (userID int, err error) {
	var hashedPassword string

	query := `SELECT id, password FROM users WHERE username = $1`
	row := conn.db.Pool.QueryRow(context.Background(), query, credentials.Username)
	err = row.Scan(&userID, &hashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			conn.Error("no such user found", http.StatusUnauthorized)
			return -1, errors.Error
		}
		log.Error.Println(err)
		conn.Error("failed to authenticate via credentials", http.StatusInternalServerError)
		return -1, errors.Error
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password))
	if err != nil {
		log.Error.Println(err)
		conn.Error("password doesn't match", http.StatusUnauthorized)
		return -1, errors.Error
	}

	return userID, nil
}

func (conn *Connection) AuthenticateViaSessionToken(sessionToken string) (userID int, err error) {
	query := `SELECT user_id FROM sessions WHERE session_token = $1`
	row := conn.db.Pool.QueryRow(context.Background(), query, sessionToken)
	err = row.Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			conn.DistinctError(
				"no such session token found",
				"unauthorized",
				http.StatusUnauthorized,
			)
			return -1, errors.Error
		}
		log.Error.Println(err)
		conn.DistinctError(
			"failed to authenticate via session token",
			"internal server error",
			http.StatusInternalServerError,
		)
		return -1, errors.Error
	}

	query = `UPDATE sessions SET last_access = NOW() WHERE session_token = $1`
	_, err = conn.db.Pool.Exec(context.Background(), query, sessionToken)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to update the last access time of the session token")
	}

	return userID, nil
}

func (conn *Connection) GenerateSessionToken(userID int) (sessionToken string, err error) {
	byteSessionToken := make([]byte, 40)
	if _, err := rand.Read(byteSessionToken); err != nil {
		log.Error.Println(err)
		conn.DistinctError(
			"failed to generate a session token",
			"internal server error",
			http.StatusInternalServerError,
		)
		return "", errors.Error
	}
	sessionToken = base64.RawURLEncoding.EncodeToString(byteSessionToken)

	query := `INSERT INTO sessions (session_token, user_id) VALUES ($1, $2)`
	_, err = conn.db.Pool.Exec(context.Background(), query, sessionToken, userID)
	if err != nil {
		log.Error.Println(err)
		conn.DistinctError(
			"failed to store a session token",
			"internal server error",
			http.StatusInternalServerError,
		)
		return "", errors.Error
	}

	return sessionToken, nil
}

func (conn *Connection) DeleteSessionToken(sessionToken string) error {
	query := `DELETE FROM sessions WHERE session_token = $1`
	_, err := conn.db.Pool.Exec(context.Background(), query, sessionToken)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to delete a session token")
		return errors.Error
	}

	return nil
}

func (conn *Connection) GetSessionToken() (string, error) {
	authHeader := conn.request.Header.Get("Authorization")
	if authHeader == "" {
		conn.writer.WriteHeader(http.StatusUnauthorized)
		return "", errors.Error
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		conn.writer.WriteHeader(http.StatusUnauthorized)
		return "", errors.Error
	}

	return parts[1], nil
}

type sessionHandler struct {
	db *db.DB
}

func (handler *sessionHandler) handlePost(conn *Connection) error {
	body, err := io.ReadAll(conn.request.Body)
	if err != nil {
		log.Error.Println(err)
		conn.Error("failed to read the request body", http.StatusBadRequest)
		return errors.Error
	}

	var credentials Credentials
	err = json.Unmarshal(body, &credentials)
	if err != nil {
		log.Error.Println(err)
		conn.Error("couldn't parse the body as a JSON object", http.StatusBadRequest)
		return errors.Error
	}

	userID, err := conn.AuthenticateViaCredentials(credentials)
	if err != nil {
		log.Error.Println("failed to authenticate via credentials")
		return err
	}

	sessionToken, err := conn.GenerateSessionToken(userID)
	if err != nil {
		log.Error.Println("failed to generate a session token")
		return err
	}

	type loginResponse struct {
		SessionToken string `json:"sessionToken"`
	}
	conn.writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(conn.writer).Encode(loginResponse{sessionToken})

	return nil
}

func (handler *sessionHandler) handleDelete(conn *Connection) error {
	sessionToken, err := conn.GetSessionToken()
	if err != nil {
		log.Error.Println("failed to get the session token")
		return errors.Error
	}

	conn.writer.WriteHeader(http.StatusOK)
	return conn.DeleteSessionToken(sessionToken)
}

func (handler *sessionHandler) handleOptions(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, OPTIONS")
	writer.WriteHeader(http.StatusOK)
}

func (handler *sessionHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if origin := request.Header.Get("Origin"); origin != "" {
		writer.Header().Set("Access-Control-Allow-Origin", origin)
	}
	writer.Header().Set("Vary", "Origin")

	conn := NewConnection(handler.db, writer, request)

	log.Info.Printf("%s %s\n", request.Method, request.URL.Path)
	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(conn); err != nil {
			log.Error.Println("failed to handle the POST request to \"/session/\"")
		}
	case http.MethodDelete:
		if err := handler.handleDelete(conn); err != nil {
			log.Error.Println("failed to handle the DELETE request to \"/session/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
