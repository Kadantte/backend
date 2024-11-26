// Copyright (C) 2024 Pavel Sobolev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"morbo/db"
	"morbo/errors"
	"morbo/log"
)

type Server struct {
	http.Server
	db *db.DB
}

func NewServer(ip string, port int) (*Server, error) {
	var server Server

	db, err := db.Prepare()
	if err != nil {
		log.Error.Println("failed to prepare the database")
		return nil, errors.Error
	}

	server.Addr = fmt.Sprintf("%s:%d", ip, port)
	server.Handler = NewServeMux(db)
	server.db = db
	return &server, nil
}

func (server *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Error.Println(err)
		log.Error.Printf("failed to listen at %s", server.Addr)
		return errors.Error
	}
	log.Info.Printf("listening at %v", server.Addr)

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
		log.Error.Printf("failed to serve: %v", err)
	}

	return server.Shutdown()
}

func (server *Server) Shutdown() error {
	log.Info.Println("shutdown initiated")
	defer log.Info.Println("shutdown finished")

	log.Info.Println("closing all database connections")
	server.db.Close()

	return server.Server.Shutdown(context.Background())
}
