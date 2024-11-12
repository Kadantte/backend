package server

import "net/http"

type ServeMux struct {
	http.ServeMux

	feeds feedsHandler
}

func NewServeMux() *ServeMux {
	mux := ServeMux{}
	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feeds/{$}", &mux.feeds)
	return &mux
}
