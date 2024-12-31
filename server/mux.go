package server

import (
	"net/http"

	"morbo/context"
	"morbo/db"
)

type baseHandler struct {
	ctx context.Context
	db  *db.DB
}

type ServeMux struct {
	http.ServeMux

	feedHandler    feedHandler
	sessionHandler sessionHandler
}

func NewServeMux(ctx context.Context, db *db.DB) *ServeMux {
	baseHandler := baseHandler{ctx, db}
	mux := ServeMux{
		feedHandler:    feedHandler{baseHandler},
		sessionHandler: sessionHandler{baseHandler},
	}

	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", &mux.feedHandler)
	mux.Handle("/session/{$}", &mux.sessionHandler)

	return &mux
}
