package server

import (
	"io"
	"net/http"
)

type feedsHandler struct{}

func (handler *feedsHandler) handlePost(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "failed to read the request body", http.StatusBadRequest)
		return
	}

	_ = body
}

func (handler *feedsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		handler.handlePost(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
