package server

import (
	"net/http"

	"morbo/db"
)

type ServeMux struct {
	http.ServeMux

	feedHandler    feedHandler
}

func NewServeMux(db *db.DB) *ServeMux {
	mux := ServeMux{
		feedHandler: feedHandler{db},
	}

	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", &mux.feedHandler)

	return &mux
}
