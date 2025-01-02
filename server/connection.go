package server

import (
	"fmt"
	"time"

	"morbo/context"
	"morbo/db"
	"morbo/log"
	"net/http"

	"github.com/google/uuid"
)

type Log struct {
	Info  log.Logger
	Error log.Logger
}

type Connection struct {
	id      string
	ctx     context.Context
	db      *db.DB
	writer  http.ResponseWriter
	request *http.Request

	cancelContext context.CancelFunc
	log           Log
}

func NewConnection(
	handler *baseHandler,
	writer http.ResponseWriter,
	request *http.Request,
) *Connection {
	ctx, cancel := context.WithTimeout(handler.ctx, 15*time.Second)
	id := uuid.New().String()
	log := Log{
		Info:  log.New(" info: ", id),
		Error: log.New("error: ", id),
	}
	conn := &Connection{id, ctx, handler.db, writer, request, cancel, log}

	context.GetWaitGroup(conn.ctx).Add(1)

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
	err := conn.ctx.Err()
	if err != nil {
		switch err {
		case context.Canceled:
			conn.Error("the request has been canceled by the server", http.StatusServiceUnavailable)
		case context.DeadlineExceed:
			conn.Error("took too long to finish the request", http.StatusGatewayTimeout)
		}
		return false
	}
	return true
}

func (conn *Connection) Disconnect() {
	conn.cancelContext()
	context.GetWaitGroup(conn.ctx).Done()
}
