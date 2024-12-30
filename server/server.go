package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"morbo/context"
	"morbo/db"
	"morbo/errors"
	"morbo/log"
)

type Server struct {
	http.Server
	db *db.DB
}

func NewServer(ctx context.Context, ip string, port int) (*Server, error) {
	var server Server

	db, err := db.Prepare(ctx)
	if err != nil {
		log.Error.Println("failed to prepare the database")
		return nil, errors.Error
	}

	server.Addr = fmt.Sprintf("%s:%d", ip, port)
	server.Handler = NewServeMux(db)
	server.db = db
	return &server, nil
}

func (server *Server) ListenAndServe(ctx context.Context) error {
	wg := context.GetWaitGroup(ctx)

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Error.Println(err)
		log.Error.Printf("failed to listen at %s", server.Addr)
		return errors.Error
	}
	log.Info.Printf("listening at %v", server.Addr)

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Info.Println("starting the server")
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Error.Println(err)
			log.Error.Println("unexpected error returned from the server")
		}
	}()

	return nil
}

func (server *Server) Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Info.Println("server shutdown initiated")
	defer log.Info.Println("server shutdown finished")

	log.Info.Println("closing all database connections")
	server.db.Close()

	if err := server.Server.Shutdown(ctx); err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to shutdown the server")
	}
}
