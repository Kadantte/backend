package server

import (
	"fmt"
	"net/http"
	"time"

	"morbo/context"
	"morbo/db"
	"morbo/log"
)

type Connection struct {
	ctx     context.Context
	db      *db.DB
	writer  http.ResponseWriter
	request *http.Request

	cancelContext context.CancelFunc
	log           log.Log
}

func BigEndianUInt40(b []byte) uint64 {
	_ = b[4]
	return uint64(b[4]) | uint64(b[3])<<8 | uint64(b[2])<<16 | uint64(b[1])<<24 | uint64(b[0])<<32
}

func NewConnection(
	handler *baseHandler,
	writer http.ResponseWriter,
	request *http.Request,
) *Connection {
	ctx, cancel := context.WithTimeout(handler.ctx, 15*time.Second)
	id := handler.newConnectionID()
	log := log.NewLog(id)
	conn := &Connection{ctx, handler.db, writer, request, cancel, log}

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
	conn.log.Error.Println(serverMessage)
	conn.writer.WriteHeader(statusCode)
	fmt.Fprint(conn.writer, userMessage)
}

func (conn *Connection) ContextAlive() bool {
	if err := conn.ctx.Err(); err != nil {
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

func (conn *Connection) Disconnect() {
	conn.cancelContext()
}
