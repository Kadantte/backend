package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"morbo/db"
)

type sessionHandler struct {
	db *db.DB
}

func (handler *sessionHandler) handlePost(writer http.ResponseWriter, request *http.Request) error {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "failed to read the request body", http.StatusBadRequest)
		return err
	}

	var credentials db.Credentials
	err = json.Unmarshal(body, &credentials)
	if err != nil {
		http.Error(writer, "couldn't parse the body as a JSON object", http.StatusBadRequest)
		return err
	}

	userID, err := handler.db.AuthenticateByCredentials(credentials)
	if errors.Is(err, db.Unathorized) {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return err
	} else if errors.Is(err, db.InternalServerError) {
		http.Error(writer, "internal server error", http.StatusInternalServerError)
		return err
	}

	sessionToken, err := handler.db.CreateSessionToken(userID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	type loginResponse struct {
		SessionToken string `json:"sessionToken"`
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(loginResponse{sessionToken})

	return nil
}

func (handler *sessionHandler) handleDelete(writer http.ResponseWriter, request *http.Request) error {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writer.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	writer.WriteHeader(http.StatusOK)

	sessionToken := parts[1]
	return handler.db.DeleteSessionToken(sessionToken)
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

	switch request.Method {
	case http.MethodPost:
		log.Println(handler.handlePost(writer, request))
	case http.MethodDelete:
		log.Println(handler.handleDelete(writer, request))
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
