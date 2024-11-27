package server

import (
	"morbo/errors"
	"morbo/log"
)

func Main(args []string) error {
	server, err := NewServer("0.0.0.0", 80)
	if err != nil {
		log.Error.Println("failed to create the server")
		return errors.Error
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Error.Println("failed to listen and serve")
		return errors.Error
	}

	return nil
}
