package server

import (
	"fmt"
	"time"

	"morbo/context"
	"morbo/db"
	"morbo/log"
	"net/http"
)

type Connection struct {
	ctx     context.Context
	db      *db.DB
	writer  http.ResponseWriter
	request *http.Request

	cancelContext context.CancelFunc
}

func NewConnection(
	handler *baseHandler,
	writer http.ResponseWriter,
	request *http.Request,
) *Connection {
	ctx, cancel := context.WithTimeout(handler.ctx, 15*time.Second)
	conn := &Connection{ctx, handler.db, writer, request, cancel}

	if origin := conn.request.Header.Get("Origin"); origin != "" {
		conn.writer.Header().Set("Access-Control-Allow-Origin", origin)
	}
	conn.writer.Header().Set("Vary", "Origin")

	return conn
}

func (conn *Connection) Error(message string, statusCode int) {
	conn.DistinctError(message, message, statusCode)
}

func (conn *Connection) DistinctError(serverMessage string, userMessage string, statusCode int) {
	log.Error.Println(serverMessage)
	conn.writer.WriteHeader(statusCode)
	fmt.Fprint(conn.writer, userMessage)
}

func (conn *Connection) Disconnect() {
	conn.cancelContext()
}
