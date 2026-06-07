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
	broker   *Broker
}

func NewServer(listener net.Listener, dir string, ctx context.Context) *Server {
	fs := http.FileServer(http.Dir(dir))

	broker := NewBroker(ctx)

	mux := http.NewServeMux()
	mux.Handle("/", LiveReloadMiddleware(fs))
	mux.Handle("/live-reload", broker)

	return &Server{
		listener: listener,
		dir:      dir,
		server: &http.Server{
			Handler:     mux,
			ReadTimeout: 15 * time.Second,
		},
		broker: broker,
	}
}

func (s *Server) Start(ctx context.Context) error {
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

func (s *Server) NotifyClients(msg []byte) {
	s.broker.Notify(msg)
}
