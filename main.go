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
	ctx, cancel := context.WithCancel(context.WithWaitGroup(context.Background()))

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, os.Interrupt)

	server, err := server.Run(ctx)
	if err != nil {
		log.Error.Println("failed to start the server")
		os.Exit(1)
	}

	select {
	case <-sigint:
		print("\r")
	}

	server.Shutdown(ctx)

	cancel()
	context.GetWaitGroup(ctx).Wait()
}
