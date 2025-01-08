package main

import (
	"os"
	"os/signal"
	"syscall"

	"morbo/context"
	"morbo/log"
	"morbo/server"
)

func main() {
	ctx, cancel := context.WithWaitGroup(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(
		sigchan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	log := log.NewLog("main")

	server, err := server.NewServer(ctx, "0.0.0.0", 80)
	if err != nil {
		log.Error.Fatalln("failed to create the server")
	}

	if err := server.ListenAndServe(ctx); err != nil {
		log.Error.Fatalln("failed to listen and serve")
	}

	<-sigchan
	print("\r")
}
