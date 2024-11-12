package server

import (
	"flag"
)

func Main(args []string) error {
	flagSet := flag.NewFlagSet("morbo", flag.ExitOnError)
	ip := flagSet.String("ip", "0.0.0.0", "ip to bind to")
	port := flagSet.Int("port", 80, "port to bind to")
	flagSet.Parse(args)

	if flagSet.NArg() > 0 {
		flagSet.Usage()
		return nil
	}

	server := NewServer(*ip, *port)
	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
