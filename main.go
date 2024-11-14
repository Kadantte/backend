package main

import (
	"fmt"
	"os"

	"morbo/server"
)

func main() {
	err := server.Main(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
