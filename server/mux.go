package server

import (
	"net/http"

	"morbo/db"
)

type ServeMux struct {
	http.ServeMux

	feeds feedsHandler
}

func NewServeMux(db *db.DB) *ServeMux {
	feeds := feedsHandler{db}
	mux := ServeMux{feeds: feeds}

	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feeds/{$}", &mux.feeds)
	return &mux
}
