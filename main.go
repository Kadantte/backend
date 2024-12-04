package main

import (
	"os"

	"morbo/context"
	"morbo/server"
)

func main() {
	err := server.Main(context.Background(), os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
