package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type feedsHandler struct {
	db *DB
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

func (handler *feedsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		handler.handlePost(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
