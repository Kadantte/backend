package main

import (
	"os"

	"morbo/server"
)

func main() {
	err := server.Main(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
