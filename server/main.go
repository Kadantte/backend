package server

import (
	"morbo/errors"
)

func Main(args []string) error {
	server, err := NewServer("0.0.0.0", 80)
	if err != nil {
		return errors.Chain("failed to create the server", err)
	}

	err = server.ListenAndServe()
	if err != nil {
		return errors.Chain("failed to listen and serve", err)
	}

	return nil
}
