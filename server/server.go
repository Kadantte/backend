package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"morbo/db"
	"morbo/errors"
)

type Server struct {
	http.Server
	db *db.DB
}

func NewServer(ip string, port int) (*Server, error) {
	var server Server

	db, err := db.Prepare()
	if err != nil {
		return nil, errors.Chain("failed to prepare the database", err)
	}

	server.Addr = fmt.Sprintf("%s:%d", ip, port)
	server.Handler = NewServeMux(db)
	server.db = db
	return &server, nil
}

func (server *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	log.Printf("listening at %v", server.Addr)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, os.Interrupt)

	errs := make(chan error, 1)
	go func() {
		errs <- server.Serve(listener)
	}()

	select {
	case <-sigint:
		print("\r")
	case err := <-errs:
		log.Printf("failed to serve: %v", err)
	}

	return server.Shutdown()
}

func (server *Server) Shutdown() error {
	log.Println("shutdown initiated")
	defer log.Println("shutdown finished")

	log.Println("closing all database connections")
	server.db.Close()

	return server.Server.Shutdown(context.Background())
}
