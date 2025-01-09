package server

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"morbo/db"
	"morbo/log"
)

type baseHandler struct {
	db  *db.DB
	log log.Log
}

func (handler *baseHandler) newConnectionID() string {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		handler.log.Error.Println("failed to generate a new ID")
	}
	number := BigEndianUInt40(bytes)
	return fmt.Sprintf("%011x", number)
}

func NewServeMux(db *db.DB) *http.ServeMux {
	baseHandler := baseHandler{db, log.NewLog("handler")}

	mux := http.ServeMux{}
	mux.Handle("/{$}", http.NotFoundHandler())
	mux.Handle("/feed/{$}", &feedHandler{baseHandler})
	mux.Handle("/session/{$}", &sessionHandler{baseHandler})

	return &mux
}
