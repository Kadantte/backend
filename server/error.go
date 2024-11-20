package server

import (
	"fmt"
	"log"
	"net/http"
)

func Error(writer http.ResponseWriter, err error, message string, statusCode int) {
	log.Println(err)
	writer.WriteHeader(statusCode)
	fmt.Fprint(writer, message)
}
