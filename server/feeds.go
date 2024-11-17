package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"morbo/db"
)

type feedsHandler struct {
	db *db.DB
}

func (handler *feedsHandler) handlePost(writer http.ResponseWriter, request *http.Request) {
	type Feed struct {
		URL string `json:"url"`
	}

	var feed Feed
	if err := json.NewDecoder(request.Body).Decode(&feed); err != nil {
		http.Error(writer, "failed to decode the request body", http.StatusBadRequest)
		return
	}

	fmt.Println(feed)
}

func (handler *feedsHandler) handleOptions(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	writer.WriteHeader(http.StatusOK)
}

func (handler *feedsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if origin := request.Header.Get("Origin"); origin != "" {
		writer.Header().Set("Access-Control-Allow-Origin", origin)
	}
	writer.Header().Set("Vary", "Origin")

	switch request.Method {
	case http.MethodPost:
		handler.handlePost(writer, request)
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
