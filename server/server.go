package server

import (
	"fmt"
	"net"
	"net/http"

	"morbo/context"
	"morbo/db"
	"morbo/errors"
	"morbo/log"
)

type Server struct {
	http.Server
	db  *db.DB
	log log.Log
}

func NewServer(ctx context.Context, ip string, port int) (*Server, error) {
	server := Server{log: log.NewLog("server")}

	db, err := db.Prepare(ctx)
	if err != nil {
		server.log.Error.Println("failed to prepare the database")
		return nil, errors.Err
	}

	server.Addr = fmt.Sprintf("%s:%d", ip, port)
	server.Handler = NewServeMux(ctx, db)
	server.db = db
	return &server, nil
}

func (server *Server) ListenAndServe(ctx context.Context) error {
	wg := context.GetWaitGroup(ctx)

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		server.log.Error.Println(err)
		server.log.Error.Printf("failed to listen at %s", server.Addr)
		return errors.Err
	}

	server.log.Info.Printf("listening at %v", server.Addr)

	wg.Add(1)
	go func() {
		server.log.Info.Println("starting the server")
		if err := server.Serve(listener); err != http.ErrServerClosed {
			server.log.Error.Println(err)
			server.log.Error.Println("unexpected error returned from the server")
		}
	}()

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
		wg.Done()
	}()

	return nil
}

func (server *Server) Shutdown(ctx context.Context) {
	server.log.Info.Println("server shutdown initiated")
	defer server.log.Info.Println("server shutdown finished")

	server.log.Info.Println("closing all database connections")
	server.db.Close()

	if err := server.Server.Shutdown(ctx); err != nil {
		server.log.Error.Println(err)
		server.log.Error.Println("failed to shutdown the server")
	}
}
