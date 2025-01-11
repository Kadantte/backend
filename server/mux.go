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

type baseHandler struct {
	db  *db.DB
	log log.Log
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

func timeoutMiddleware(handler http.Handler) http.Handler {
	f := func(writer http.ResponseWriter, request *http.Request) {
		timeoutWriter := timeoutResponseWriter{ResponseWriter: writer}

		done := make(chan struct{})
		go func() {
			handler.ServeHTTP(&timeoutWriter, request)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(15 * time.Second):
			timeoutWriter.timedOut.Store(true)

			writer.WriteHeader(http.StatusGatewayTimeout)
			writer.Write([]byte("took too long to finish the request"))
		}
	}

	return http.HandlerFunc(f)
}

func (handler *baseHandler) newConnectionID() string {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		handler.log.Error.Println("failed to generate a new ID")
	}
	number := BigEndianUInt40(bytes)
	return fmt.Sprintf("%011x", number)
}

func NewServeMux(db *db.DB) *http.ServeMux {
	baseHandler := baseHandler{db, log.NewLog("handler")}
	feedHandler := timeoutMiddleware(&feedHandler{baseHandler})
	sessionHandler := timeoutMiddleware(&sessionHandler{baseHandler})

	mux := http.ServeMux{}
	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", feedHandler)
	mux.Handle("/session/{$}", sessionHandler)

	return &mux
}
