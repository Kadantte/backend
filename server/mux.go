package server

import (
	"net/http"

	"morbo/db"
)

type ServeMux struct {
	http.ServeMux

	feedHandler    feedHandler
	sessionHandler sessionHandler
}

func NewServeMux(db *db.DB) *ServeMux {
	mux := ServeMux{
		feedHandler:    feedHandler{db},
		sessionHandler: sessionHandler{db},
	}

	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", &mux.feedHandler)
	mux.Handle("/session/{$}", &mux.feedHandler)

	return &mux
}
