package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"morbo/context"
	"morbo/db"
	"morbo/errors"
	"morbo/log"

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
	row := conn.db.Pool.QueryRow(ctx, query, credentials.Username)
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

func (conn *Connection) AuthenticateViaSessionToken(ctx context.Context, sessionToken string) (userID int, err error) {
	query := `SELECT user_id FROM sessions WHERE session_token = $1`
	row := conn.db.Pool.QueryRow(ctx, query, sessionToken)
	err = row.Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
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
	_, err = conn.db.Pool.Exec(ctx, query, sessionToken)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to update the last access time of the session token")
	}

	return userID, nil
}

func (conn *Connection) GenerateSessionToken(ctx context.Context, userID int) (sessionToken string, err error) {
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
	_, err = conn.db.Pool.Exec(ctx, query, sessionToken, userID)
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

func (conn *Connection) DeleteSessionToken(ctx context.Context, sessionToken string) error {
	query := `DELETE FROM sessions WHERE session_token = $1`
	_, err := conn.db.Pool.Exec(ctx, query, sessionToken)
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
	db *db.DB
}

func (handler *sessionHandler) handlePost(ctx context.Context, conn *Connection) error {
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

	userID, err := conn.AuthenticateViaCredentials(ctx, credentials)
	if err != nil {
		log.Error.Println("failed to authenticate via credentials")
		return err
	}

	sessionToken, err := conn.GenerateSessionToken(ctx, userID)
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

func (handler *sessionHandler) handleDelete(ctx context.Context, conn *Connection) error {
	sessionToken, err := conn.GetSessionToken()
	if err != nil {
		log.Error.Println("failed to get the session token")
		return errors.Error
	}

	conn.writer.WriteHeader(http.StatusOK)

	err = conn.DeleteSessionToken(ctx, sessionToken)
	if err != nil {
		log.Error.Println("failed to delete the session token")
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn := NewConnection(handler.db, writer, request)

	log.Info.Printf("%s %s\n", request.Method, request.URL.Path)
	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(ctx, conn); err != nil {
			log.Error.Println("failed to handle the POST request to \"/session/\"")
		}
	case http.MethodDelete:
		if err := handler.handleDelete(ctx, conn); err != nil {
			log.Error.Println("failed to handle the DELETE request to \"/session/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
