package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phmshk/go-live-server/internal/server"
	"github.com/phmshk/go-live-server/internal/utils"
)

func main() {
	defaultPort := 8080
	flagPort := flag.Int("port", defaultPort, "starting port number")
	flagDir := flag.String("dir", ".", "directory to serve")

	flag.Parse()

	listener, err := utils.GetAvailablePort(*flagPort)
	if err != nil {
		log.Fatalf("failed to find an available port: %v", err)
	}
	defer listener.Close()

	s := server.NewServer(listener, *flagDir)

	shutDownSignal := make(chan os.Signal, 1)
	signal.Notify(shutDownSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Printf("Server stopped unexpectedly: %v", err)
		}
	}()

	<-shutDownSignal

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server successfully stopped. Bye!")
}
