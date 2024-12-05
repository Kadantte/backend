package server

import (
	"morbo/context"
	"morbo/errors"
	"morbo/log"
)

func Run(ctx context.Context) (*Server, error) {
	server, err := NewServer("0.0.0.0", 80)
	if err != nil {
		log.Error.Println("failed to create the server")
		return nil, errors.Error
	}

	err = server.ListenAndServe(ctx)
	if err != nil {
		log.Error.Println("failed to listen and serve")
		return nil, errors.Error
	}

	return server, nil
}
