package server

import (
	"morbo/context"
	"morbo/errors"
	"morbo/log"
)

func Run(ctx context.Context, log *log.Log) (*Server, error) {
	server, err := NewServer(ctx, "0.0.0.0", 80)
	if err != nil {
		log.Error.Println("failed to create the server")
		return nil, errors.Err
	}

	if err = server.ListenAndServe(ctx); err != nil {
		server.log.Error.Println("failed to listen and serve")
		return nil, errors.Err
	}

	return server, nil
}
