package server

import (
	"net/http"

	"morbo/context"
	"morbo/db"
)

type ServeMux struct {
	http.ServeMux

	feedHandler    feedHandler
	sessionHandler sessionHandler
}

func NewServeMux(ctx context.Context, db *db.DB) *ServeMux {
	mux := ServeMux{
		feedHandler:    feedHandler{ctx, db},
		sessionHandler: sessionHandler{ctx, db},
	}

	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", &mux.feedHandler)
	mux.Handle("/session/{$}", &mux.sessionHandler)

	return &mux
}
