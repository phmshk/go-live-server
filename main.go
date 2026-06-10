package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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

type Config struct {
	Port int
	Dir  string
}

const (
	defaultPort = 8080
	defaultDir  = "."
)

func main() {
	cfg := &Config{}

	flagPort := flag.Int("port", defaultPort, "starting port number")
	flagDir := flag.String("dir", defaultDir, "directory to serve")
	flag.Parse()

	cfg.Port = *flagPort
	cfg.Dir = *flagDir

	if err := Run(cfg); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func Run(cfg *Config) error {
	ctxWatcher, cancelWatcherCtx := context.WithCancel(context.Background())
	defer cancelWatcherCtx()

	listener, err := utils.GetAvailablePort(cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to find an available port: %w", err)
	}
	defer listener.Close()

	w, err := watcher.NewWatcher(cfg.Dir, 100*time.Millisecond)
	if err != nil {
		return fmt.Errorf("error creating watcher: %w", err)
	}

	s := server.NewServer(listener, cfg.Dir, ctxWatcher)

	go w.Start(ctxWatcher, func(fileName string) {
		s.NotifyClients([]byte(fileName))
	})

	shutDownSignal := make(chan os.Signal, 1)
	signal.Notify(shutDownSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.Start(ctxWatcher); err != nil {
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

	return nil
}
