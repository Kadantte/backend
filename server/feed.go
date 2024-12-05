package server

import (
	"encoding/json"
	"net/http"

	"morbo/db"
	"morbo/errors"
	"morbo/log"
)

type feedHandler struct {
	db *db.DB
}

func (handler *feedHandler) handlePost(writer http.ResponseWriter, request *http.Request) error {
	sessionToken, err := GetSessionToken(writer, request)
	if err == errors.Done {
		return nil
	}

	_, statusCode, err := handler.db.AuthenticateBySessionToken(sessionToken)
	if err != nil {
		Error(writer, "failed to authenticate the user", statusCode)
		return errors.Error
	}

	type RequestBody struct {
		URL string `json:"url"`
	}

	var requestBody RequestBody
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		log.Error.Println(err)
		Error(writer, "failed to decode the request body", http.StatusBadRequest)
		return errors.Error
	}

	rss, statusCode, err := parseRSS(requestBody.URL)
	if err != nil {
		Error(writer, "failed to parse the RSS feed", statusCode)
		return errors.Error
	}

	type ResponseBody struct {
		Title string `json:"title"`
	}
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(ResponseBody{rss.Channel.Title})

	return nil
}

func (handler *feedHandler) handleOptions(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	writer.WriteHeader(http.StatusOK)
}

func (handler *feedHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if origin := request.Header.Get("Origin"); origin != "" {
		writer.Header().Set("Access-Control-Allow-Origin", origin)
	}
	writer.Header().Set("Vary", "Origin")

	log.Info.Printf("%s %s %s\n", request.RemoteAddr, request.Method, request.URL.Path)
	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(writer, request); err != nil {
			log.Error.Println("failed to handle the POST request to \"/feed/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
