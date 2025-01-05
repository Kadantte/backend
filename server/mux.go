package server

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"morbo/context"
	"morbo/db"
	"morbo/log"
)

type baseHandler struct {
	ctx context.Context
	db  *db.DB
	log log.Log
}

func (handler *baseHandler) newConnectionID() string {
	bytes := make([]byte, 5)
	_, err := rand.Read(bytes)
	if err != nil {
		handler.log.Error.Println("failed to generate a new ID")
	}
	number := BigEndianUInt40(bytes)
	return fmt.Sprintf("%011x", number)
}

type ServeMux struct {
	http.ServeMux

	feedHandler    feedHandler
	sessionHandler sessionHandler
}

func NewServeMux(ctx context.Context, db *db.DB) *ServeMux {
	baseHandler := baseHandler{ctx, db, log.NewLog("handler")}
	mux := ServeMux{
		feedHandler:    feedHandler{baseHandler},
		sessionHandler: sessionHandler{baseHandler},
	}

	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", &mux.feedHandler)
	mux.Handle("/session/{$}", &mux.sessionHandler)

	return &mux
}
