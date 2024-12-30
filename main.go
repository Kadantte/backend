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

	sigchan := make(chan os.Signal, 1)
	signal.Notify(
		sigchan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	server, err := server.Run(ctx)
	if err != nil {
		log.Error.Println("failed to start the server")
		os.Exit(1)
	}

	<-sigchan
	print("\r")

	server.Shutdown(ctx)

	cancel()
	context.GetWaitGroup(ctx).Wait()
}
