package server

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"morbo/db"
	"morbo/log"
)

type handlerCtx struct {
	db  *db.DB
	log log.Log
}

func BigEndianUInt40(b []byte) uint64 {
	_ = b[4]
	return uint64(b[4]) | uint64(b[3])<<8 | uint64(b[2])<<16 | uint64(b[1])<<24 | uint64(b[0])<<32
}

func (handlerCtx *handlerCtx) newConnectionID() string {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		handlerCtx.log.Error.Println("failed to generate a new ID")
	}
	number := BigEndianUInt40(bytes)
	return fmt.Sprintf("%011x", number)
}

func (handlerCtx *handlerCtx) newConnection(writer http.ResponseWriter, request *http.Request) *Connection {
	id := handlerCtx.newConnectionID()
	log := log.NewLog(id)
	return &Connection{handlerCtx.db, &log, writer, request}
}

type HandlerFunc func(*Connection)

func (f HandlerFunc) Handle(conn *Connection) {
	f(conn)
}

type Handler interface {
	Handle(*Connection)
}

type timeoutResponseWriter struct {
	http.ResponseWriter
	timedOut atomic.Bool
}

func (timeoutWriter *timeoutResponseWriter) WriteHeader(statusCode int) {
	if timeoutWriter.timedOut.Load() {
		return
	}

	timeoutWriter.ResponseWriter.WriteHeader(statusCode)
}

func (timeoutWriter *timeoutResponseWriter) Write(b []byte) (int, error) {
	if timeoutWriter.timedOut.Load() {
		return 0, http.ErrHandlerTimeout
	}

	return timeoutWriter.ResponseWriter.Write(b)
}

func timeoutMiddleware(handler Handler) Handler {
	f := func(conn *Connection) {
		timeoutWriter := timeoutResponseWriter{ResponseWriter: conn.writer}

		timeoutConn := *conn
		timeoutConn.writer = &timeoutWriter

		done := make(chan struct{})
		go func() {
			handler.Handle(&timeoutConn)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(15 * time.Second):
			timeoutWriter.timedOut.Store(true)
			conn.Error("took too long to finish the request", http.StatusGatewayTimeout)
		}
	}

	return HandlerFunc(f)
}

func finalHandler(handlerCtx *handlerCtx, handler Handler) http.Handler {
	finalHandler := timeoutMiddleware(handler)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		conn := handlerCtx.newConnection(writer, request)
		conn.SendOriginHeaders()

		finalHandler.Handle(conn)
	})
}

func NewServeMux(db *db.DB) *http.ServeMux {
	handlerCtx := handlerCtx{db, log.NewLog("handler")}
	feedHandler := finalHandler(&handlerCtx, &feedHandler{})
	sessionHandler := finalHandler(&handlerCtx, &sessionHandler{})

	mux := http.ServeMux{}
	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", feedHandler)
	mux.Handle("/session/{$}", sessionHandler)

	return &mux
}
