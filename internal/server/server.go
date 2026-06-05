package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type Server struct {
	dir      string
	listener net.Listener
	server   *http.Server
}

func NewServer(listener net.Listener, dir string) *Server {
	fs := http.FileServer(http.Dir(dir))
	mux := http.NewServeMux()
	mux.Handle("/", fs)

	return &Server{
		listener: listener,
		dir:      dir,
		server: &http.Server{
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	_, port, err := net.SplitHostPort(s.listener.Addr().String())
	if err != nil {
		port = s.listener.Addr().String()
	}

	log.Printf("Starting server on http://localhost:%s serving %s", port, s.dir)

	err = s.server.Serve(s.listener)
	if err != nil {
		return fmt.Errorf("error starting file server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
