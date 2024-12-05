package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"morbo/db"
	"morbo/errors"
	"morbo/log"
)

func GetSessionToken(writer http.ResponseWriter, request *http.Request) (string, error) {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		return "", errors.Done
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writer.WriteHeader(http.StatusUnauthorized)
		return "", errors.Done
	}

	return parts[1], nil
}

type sessionHandler struct {
	db *db.DB
}

func (handler *sessionHandler) handlePost(writer http.ResponseWriter, request *http.Request) error {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Error.Println(err)
		Error(writer, "failed to read the request body", http.StatusBadRequest)
		return errors.Error
	}

	var credentials db.Credentials
	err = json.Unmarshal(body, &credentials)
	if err != nil {
		log.Error.Println(err)
		Error(writer, "couldn't parse the body as a JSON object", http.StatusBadRequest)
		return errors.Error
	}

	userID, statusCode, err := handler.db.AuthenticateByCredentials(credentials)
	if err != nil {
		switch statusCode {
		case http.StatusUnauthorized:
			Error(writer, "unauthorized", http.StatusUnauthorized)
			return errors.Error
		case http.StatusInternalServerError:
			Error(writer, "internal server error", http.StatusInternalServerError)
			return errors.Error
		}
	}

	sessionToken, err := handler.db.GenerateSessionToken(userID)
	if err != nil {
		Error(writer, "failed to generate a session token", http.StatusInternalServerError)
		return errors.Error
	}

	type loginResponse struct {
		SessionToken string `json:"sessionToken"`
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(loginResponse{sessionToken})

	return nil
}

func (handler *sessionHandler) handleDelete(writer http.ResponseWriter, request *http.Request) error {
	sessionToken, err := GetSessionToken(writer, request)
	if err == errors.Done {
		return nil
	}

	writer.WriteHeader(http.StatusOK)
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

	log.Info.Printf("%s %s\n", request.Method, request.URL.Path)
	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(writer, request); err != nil {
			log.Error.Println("failed to handle the POST request to \"/session/\"")
		}
	case http.MethodDelete:
		if err := handler.handleDelete(writer, request); err != nil {
			log.Error.Println("failed to handle the DELETE request to \"/session/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
