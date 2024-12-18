package server

import (
	"fmt"
	"morbo/db"
	"morbo/log"
	"net/http"
)

type Connection struct {
	db      *db.DB
	writer  http.ResponseWriter
	request *http.Request
}

func NewConnection(db *db.DB, writer http.ResponseWriter, request *http.Request) *Connection {
	return &Connection{db, writer, request}
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (conn *Connection) Error(message string, statusCode int) {
	conn.DistinctError(message, message, statusCode)
}

func (conn *Connection) DistinctError(serverMessage string, userMessage string, statusCode int) {
	log.Error.Println(serverMessage)
	conn.writer.WriteHeader(statusCode)
	fmt.Fprint(conn.writer, userMessage)
}
