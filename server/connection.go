package server

import (
	"crypto/rand"
	"fmt"
	"time"

	"morbo/context"
	"morbo/db"
	"morbo/log"
	"net/http"
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

func BigEndianUInt40(b []byte) uint64 {
	_ = b[4]
	return uint64(b[4]) | uint64(b[3])<<8 | uint64(b[2])<<16 | uint64(b[1])<<24 | uint64(b[0])<<32
}

func newID() string {
	bytes := make([]byte, 5)
	rand.Read(bytes)
	number := BigEndianUInt40(bytes)
	return fmt.Sprintf("%013x", number)
}

func NewConnection(
	handler *baseHandler,
	writer http.ResponseWriter,
	request *http.Request,
) *Connection {
	ctx, cancel := context.WithTimeout(handler.ctx, 15*time.Second)
	id := newID()
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
