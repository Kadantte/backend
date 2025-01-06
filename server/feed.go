package server

import (
	"bytes"
	"encoding/json"
	"net/http"

	"morbo/errors"
)

type feedHandler struct {
	baseHandler
}

func (handler *feedHandler) handlePost(conn *Connection) error {
	sessionToken, err := conn.GetSessionToken()
	if err != nil {
		conn.log.Error.Println("failed to get the session token")
		return errors.Error
	}

	if _, err := conn.AuthenticateViaSessionToken(sessionToken); err != nil {
		conn.log.Error.Println("failed to authenticate via the session token")
		return errors.Error
	}

	type RequestBody struct {
		URL string `json:"url"`
	}

	var requestBody RequestBody
	if err := json.NewDecoder(conn.request.Body).Decode(&requestBody); err != nil {
		conn.log.Error.Println(err)
		conn.Error("failed to decode the request body", http.StatusBadRequest)
		return errors.Error
	}

	rss, err := conn.parseRSS(requestBody.URL)
	if err != nil {
		conn.log.Error.Println("failed to parse the RSS feed")
		return errors.Error
	}

	type ResponseBody struct {
		Title string `json:"title"`
	}

	responseBody := ResponseBody{rss.Channel.Title}

	var responseBodyBuffer bytes.Buffer
	if err := json.NewEncoder(&responseBodyBuffer).Encode(&responseBody); err != nil {
		conn.DistinctError(
			"failed to encode the response",
			"internal server error",
			http.StatusInternalServerError,
		)
		return errors.Error
	}

	conn.writer.Header().Set("Content-Type", "application/json")

	if _, err = responseBodyBuffer.WriteTo(conn.writer); err != nil {
		conn.log.Error.Println("failed to write to the body")
		return errors.Error
	}

	return nil
}

func (handler *feedHandler) handleOptions(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	writer.WriteHeader(http.StatusOK)
}

func (handler *feedHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn := NewConnection(&handler.baseHandler, writer, request)
	defer conn.Disconnect()

	conn.log.Info.Printf("%s %s %s\n", request.RemoteAddr, request.Method, request.URL.Path)

	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(conn); err != nil {
			conn.log.Error.Println("failed to handle the POST request to \"/feed/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
