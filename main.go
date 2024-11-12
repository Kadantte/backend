package main

import (
	"log"
	"os"

	"morbo/server"
)

func main() {
	err := server.Main(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
