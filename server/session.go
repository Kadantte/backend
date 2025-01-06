package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"morbo/errors"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (conn *Connection) AuthenticateViaCredentials(credentials Credentials) (userID int, err error) {
	if !conn.ContextAlive() {
		return -1, errors.Error
	}

	var hashedPassword string

	query := `SELECT id, password FROM users WHERE username = $1`
	row := conn.QueryRow(query, credentials.Username)
	if err = conn.ScanRow(row, &userID, &hashedPassword); err != nil {
		switch err {
		case pgx.ErrNoRows:
			conn.Error("no such user found", http.StatusUnauthorized)
		default:
			conn.log.Error.Println("failed to retrieve the stored credentials")
		}
		return -1, errors.Error
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password)); err != nil {
		conn.log.Error.Println(err)
		conn.Error("password doesn't match", http.StatusUnauthorized)
		return -1, errors.Error
	}

	return userID, nil
}

func (conn *Connection) AuthenticateViaSessionToken(sessionToken string) (userID int, err error) {
	if !conn.ContextAlive() {
		return -1, errors.Error
	}

	query := `SELECT user_id FROM sessions WHERE session_token = $1`
	row := conn.QueryRow(query, sessionToken)
	if err = conn.ScanRow(row, &userID); err != nil {
		switch err {
		case pgx.ErrNoRows:
			conn.Error("no such session token found", http.StatusUnauthorized)
		default:
			conn.log.Error.Println("failed to retrieve the session token")
		}
		return -1, errors.Error
	}

	query = `UPDATE sessions SET last_access = NOW() WHERE session_token = $1`
	if err := conn.Exec(query, sessionToken); err != nil {
		conn.log.Error.Println("failed to update the last access time of the session token")
		return -1, errors.Error
	}

	return userID, nil
}

func (conn *Connection) GenerateSessionToken(userID int) (sessionToken string, err error) {
	if !conn.ContextAlive() {
		return "", errors.Error
	}

	byteSessionToken := make([]byte, 40)
	if _, err := rand.Read(byteSessionToken); err != nil {
		conn.log.Error.Println(err)
		conn.DistinctError(
			"failed to generate a session token",
			"internal server error",
			http.StatusInternalServerError,
		)
		return "", errors.Error
	}

	sessionToken = base64.RawURLEncoding.EncodeToString(byteSessionToken)

	query := `INSERT INTO sessions (session_token, user_id) VALUES ($1, $2)`
	if err := conn.Exec(query, sessionToken, userID); err != nil {
		conn.log.Error.Println("failed to store a session token")
		return "", errors.Error
	}

	return sessionToken, nil
}

func (conn *Connection) DeleteSessionToken(sessionToken string) error {
	if !conn.ContextAlive() {
		return errors.Error
	}

	query := `DELETE FROM sessions WHERE session_token = $1`
	if err := conn.Exec(query, sessionToken); err != nil {
		conn.log.Error.Println("failed to execute the statement for deleting the session token")
		return errors.Error
	}

	return nil
}

func (conn *Connection) GetSessionToken() (string, error) {
	if !conn.ContextAlive() {
		return "", errors.Error
	}

	authHeader := conn.request.Header.Get("Authorization")
	if authHeader == "" {
		conn.Error("empty Authorization header", http.StatusUnauthorized)
		return "", errors.Error
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		conn.Error("invalid Authorization header", http.StatusUnauthorized)
		return "", errors.Error
	}

	return parts[1], nil
}

type sessionHandler struct {
	baseHandler
}

func (handler *sessionHandler) handlePost(conn *Connection) error {
	type RequestBody = Credentials

	var requestBody RequestBody
	if err := json.NewDecoder(conn.request.Body).Decode(&requestBody); err != nil {
		conn.log.Error.Println(err)
		conn.Error("failed to decode the request body", http.StatusBadRequest)
		return errors.Error
	}

	userID, err := conn.AuthenticateViaCredentials(requestBody)
	if err != nil {
		conn.log.Error.Println("failed to authenticate via credentials")
		return err
	}

	sessionToken, err := conn.GenerateSessionToken(userID)
	if err != nil {
		conn.log.Error.Println("failed to generate a session token")
		return err
	}

	type ResponseBody struct {
		SessionToken string `json:"sessionToken"`
	}

	responseBody := ResponseBody{sessionToken}

	var responseBodyBuffer bytes.Buffer
	if err := json.NewEncoder(&responseBodyBuffer).Encode(responseBody); err != nil {
		conn.DistinctError(
			"failed to encode the response",
			"internal server error",
			http.StatusInternalServerError,
		)
		return errors.Error
	}

	conn.writer.Header().Set("Content-Type", "application/json")

	if _, err := responseBodyBuffer.WriteTo(conn.writer); err != nil {
		conn.log.Error.Println("failed to write to the body")
		return errors.Error
	}

	return nil
}

func (handler *sessionHandler) handleDelete(conn *Connection) error {
	sessionToken, err := conn.GetSessionToken()
	if err != nil {
		conn.log.Error.Println("failed to get the session token")
		return errors.Error
	}

	conn.writer.WriteHeader(http.StatusOK)

	if err := conn.DeleteSessionToken(sessionToken); err != nil {
		conn.log.Error.Println("failed to delete the session token")
		return errors.Error
	}

	return nil
}

func (handler *sessionHandler) handleOptions(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, OPTIONS")
	writer.WriteHeader(http.StatusOK)
}

func (handler *sessionHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn := NewConnection(&handler.baseHandler, writer, request)
	defer conn.Disconnect()

	conn.log.Info.Printf("%s %s\n", request.Method, request.URL.Path)

	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(conn); err != nil {
			conn.log.Error.Println("failed to handle the POST request to \"/session/\"")
		}
	case http.MethodDelete:
		if err := handler.handleDelete(conn); err != nil {
			conn.log.Error.Println("failed to handle the DELETE request to \"/session/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
