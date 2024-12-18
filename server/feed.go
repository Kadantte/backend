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

func (handler *feedHandler) handlePost(conn *Connection) error {
	sessionToken, err := conn.GetSessionToken()
	if err == errors.Done {
		log.Error.Println("failed to get the session token")
		return nil
	}

	_, err = conn.AuthenticateViaSessionToken(sessionToken)
	if err == errors.Done {
		log.Error.Println("failed to authenticate by the session token")
		return err
	}

	type RequestBody struct {
		URL string `json:"url"`
	}

	var requestBody RequestBody
	if err := json.NewDecoder(conn.request.Body).Decode(&requestBody); err != nil {
		log.Error.Println(err)
		conn.Error("failed to decode the request body", http.StatusBadRequest)
		return errors.Error
	}

	rss, err := conn.parseRSS(requestBody.URL)
	if err != nil {
		log.Error.Println("failed to parse the RSS feed")
		return errors.Error
	}

	type ResponseBody struct {
		Title string `json:"title"`
	}
	conn.writer.WriteHeader(http.StatusOK)
	json.NewEncoder(conn.writer).Encode(ResponseBody{rss.Channel.Title})

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

	conn := NewConnection(handler.db, writer, request)

	log.Info.Printf("%s %s %s\n", request.RemoteAddr, request.Method, request.URL.Path)
	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(conn); err != nil {
			log.Error.Println("failed to handle the POST request to \"/feed/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
