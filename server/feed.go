package server

import (
	"encoding/json"
	"log"
	"net/http"

	"morbo/db"
)

type feedHandler struct {
	db *db.DB
}

func (handler *feedHandler) handlePost(writer http.ResponseWriter, request *http.Request) {
	type RequestBody struct {
		URL string `json:"url"`
	}

	var requestBody RequestBody
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		log.Println(err)
		http.Error(writer, "failed to decode the request body", http.StatusBadRequest)
		return
	}

	rss, err := parseRSS(requestBody.URL)
	if err != nil {
		log.Println(err)
		http.Error(writer, "failed to parse the RSS feed", http.StatusBadRequest)
		return
	}

	type ResponseBody struct {
		Title string `json:"title"`
	}
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(ResponseBody{rss.Channel.Title})
}

func (handler *feedHandler) handleOptions(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	writer.WriteHeader(http.StatusOK)
}

func (handler *feedHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
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
