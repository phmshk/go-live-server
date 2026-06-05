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
	"github.com/phmshk/go-live-server/internal/watcher"
)

func main() {
	defaultPort := 8080
	flagPort := flag.Int("port", defaultPort, "starting port number")
	flagDir := flag.String("dir", ".", "directory to serve")

	flag.Parse()

	ctxWatcher, cancelWatcherCtx := context.WithCancel(context.Background())
	defer cancelWatcherCtx()

	listener, err := utils.GetAvailablePort(*flagPort)
	if err != nil {
		log.Fatalf("failed to find an available port: %v", err)
	}
	defer listener.Close()

	w, err := watcher.NewWatcher(*flagDir)
	if err != nil {
		log.Fatalf("error creating watcher: %v", err)
	}

	s := server.NewServer(listener, *flagDir)

	go w.Start(ctxWatcher, func(fileName string) { s.NotifyClients([]byte("reload")) })

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
	cancelWatcherCtx()

	log.Println("Shutting down...")

	ctxShutdown, cancelShutdownCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdownCtx()

	if err := s.Shutdown(ctxShutdown); err != nil {
		log.Printf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server successfully stopped. Bye!")
}
