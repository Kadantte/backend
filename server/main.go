package server

func Main(args []string) error {
	server, err := NewServer("0.0.0.0", 80)
	if err != nil {
		return err
	}

	err = server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
