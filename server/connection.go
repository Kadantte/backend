package server

import (
	"net/http"

	"morbo/context"
	"morbo/db"
	"morbo/log"
)

type Connection struct {
	db      *db.DB
	log     *log.Log
	writer  http.ResponseWriter
	request *http.Request
}

func (conn *Connection) SendOriginHeaders() {
	if origin := conn.request.Header.Get("Origin"); origin != "" {
		conn.writer.Header().Set("Access-Control-Allow-Origin", origin)
	}
	conn.writer.Header().Set("Vary", "Origin")
}

func (conn *Connection) Error(message string, statusCode int) {
	conn.DistinctError(message, message, statusCode)
}

func (conn *Connection) DistinctError(serverMessage string, userMessage string, statusCode int) {
	conn.log.Error.Println(serverMessage)
	conn.writer.WriteHeader(statusCode)
	if _, err := conn.writer.Write([]byte(userMessage)); err != nil {
		conn.log.Error.Println("failed to write the response")
	}
}

func (conn *Connection) ContextAlive(ctx context.Context) bool {
	if err := ctx.Err(); err != nil {
		switch err {
		case context.ErrCanceled:
			conn.Error("the request has been canceled by the server", http.StatusServiceUnavailable)
		case context.ErrDeadlineExceed:
			conn.Error("took too long to finish the request", http.StatusGatewayTimeout)
		}
		return false
	}
	return true
}
