package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"morbo/context"
	"morbo/errors"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (conn *Connection) AuthenticateViaCredentials(ctx context.Context, credentials Credentials) (userID int, err error) {
	var hashedPassword string

	query := `SELECT id, password FROM users WHERE username = $1`
	row := conn.QueryRow(ctx, query, credentials.Username)
	if err = conn.ScanRow(ctx, row, &userID, &hashedPassword); err != nil {
		switch err {
		case pgx.ErrNoRows:
			conn.Error("no such user found", http.StatusUnauthorized)
		default:
			conn.log.Error.Println("failed to retrieve the stored credentials")
		}
		return -1, errors.Err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password)); err != nil {
		conn.log.Error.Println(err)
		conn.Error("password doesn't match", http.StatusUnauthorized)
		return -1, errors.Err
	}

	return userID, nil
}

func (conn *Connection) AuthenticateViaSessionToken(ctx context.Context, sessionToken string) (userID int, err error) {
	query := `SELECT user_id FROM sessions WHERE session_token = $1`
	row := conn.QueryRow(ctx, query, sessionToken)
	if err = conn.ScanRow(ctx, row, &userID); err != nil {
		switch err {
		case pgx.ErrNoRows:
			conn.Error("no such session token found", http.StatusUnauthorized)
		default:
			conn.log.Error.Println("failed to retrieve the session token")
		}
		return -1, errors.Err
	}

	query = `UPDATE sessions SET last_access = NOW() WHERE session_token = $1`
	if err := conn.Exec(ctx, query, sessionToken); err != nil {
		conn.log.Error.Println("failed to update the last access time of the session token")
		return -1, errors.Err
	}

	return userID, nil
}

func (conn *Connection) GenerateSessionToken(ctx context.Context, userID int) (sessionToken string, err error) {
	byteSessionToken := make([]byte, 40)
	if _, err := rand.Read(byteSessionToken); err != nil {
		conn.log.Error.Println(err)
		conn.DistinctError(
			"failed to generate a session token",
			"internal server error",
			http.StatusInternalServerError,
		)
		return "", errors.Err
	}

	sessionToken = base64.RawURLEncoding.EncodeToString(byteSessionToken)

	query := `INSERT INTO sessions (session_token, user_id) VALUES ($1, $2)`
	if err := conn.Exec(ctx, query, sessionToken, userID); err != nil {
		conn.log.Error.Println("failed to store a session token")
		return "", errors.Err
	}

	return sessionToken, nil
}

func (conn *Connection) DeleteSessionToken(ctx context.Context, sessionToken string) error {
	query := `DELETE FROM sessions WHERE session_token = $1`
	if err := conn.Exec(ctx, query, sessionToken); err != nil {
		conn.log.Error.Println("failed to execute the statement for deleting the session token")
		return errors.Err
	}

	return nil
}

func (conn *Connection) GetSessionToken() (string, error) {
	authHeader := conn.request.Header.Get("Authorization")
	if authHeader == "" {
		conn.Error("empty Authorization header", http.StatusUnauthorized)
		return "", errors.Err
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		conn.Error("invalid Authorization header", http.StatusUnauthorized)
		return "", errors.Err
	}

	return parts[1], nil
}

type sessionHandler struct{}

func (handler *sessionHandler) handlePost(ctx context.Context, conn *Connection) error {
	type RequestBody = Credentials

	var requestBody RequestBody
	if err := json.NewDecoder(conn.request.Body).Decode(&requestBody); err != nil {
		conn.log.Error.Println(err)
		conn.Error("failed to decode the request body", http.StatusBadRequest)
		return errors.Err
	}

	userID, err := conn.AuthenticateViaCredentials(ctx, requestBody)
	if err != nil {
		conn.log.Error.Println("failed to authenticate via credentials")
		return err
	}

	sessionToken, err := conn.GenerateSessionToken(ctx, userID)
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
		return errors.Err
	}

	conn.writer.Header().Set("Content-Type", "application/json")

	if _, err := responseBodyBuffer.WriteTo(conn.writer); err != nil {
		conn.log.Error.Println("failed to write to the body")
		return errors.Err
	}

	return nil
}

func (handler *sessionHandler) handleDelete(ctx context.Context, conn *Connection) error {
	sessionToken, err := conn.GetSessionToken()
	if err != nil {
		conn.log.Error.Println("failed to get the session token")
		return errors.Err
	}

	conn.writer.WriteHeader(http.StatusOK)

	if err := conn.DeleteSessionToken(ctx, sessionToken); err != nil {
		conn.log.Error.Println("failed to delete the session token")
		return errors.Err
	}

	return nil
}

func (handler *sessionHandler) handleOptions(conn *Connection) {
	conn.writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	conn.writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, OPTIONS")
	conn.writer.WriteHeader(http.StatusOK)
}

func (handler *sessionHandler) Handle(conn *Connection) {
	ctx := conn.request.Context()

	conn.log.Info.Printf(
		"%s %s %s\n",
		conn.request.RemoteAddr,
		conn.request.Method,
		conn.request.URL.Path,
	)

	switch conn.request.Method {
	case http.MethodPost:
		if err := handler.handlePost(ctx, conn); err != nil {
			conn.log.Error.Println("failed to handle the POST request to \"/session/\"")
		}
	case http.MethodDelete:
		if err := handler.handleDelete(ctx, conn); err != nil {
			conn.log.Error.Println("failed to handle the DELETE request to \"/session/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(conn)
	default:
		conn.writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
