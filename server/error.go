package server

import (
	"fmt"
	"net/http"

	"morbo/log"
)

func Error(writer http.ResponseWriter, message string, statusCode int) {
	log.Error.Println(message)
	writer.WriteHeader(statusCode)
	fmt.Fprint(writer, message)
}
