package server

import (
	"fmt"
	"net/http"
)

type Broker struct {
	Notifier       chan []byte              // channel for all recievers
	newClients     chan chan []byte         // channel to register new recievers
	closingClients chan chan []byte         // channel to delete inactive recievers
	clients        map[chan []byte]struct{} // map of active clients
}

func NewBroker() *Broker {
	b := &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]struct{}),
	}

	return b
}

func (b *Broker) Listen() {
	for {
		select {
		case newClient := <-b.newClients:
			b.clients[newClient] = struct{}{}

		case closingClient := <-b.closingClients:
			delete(b.clients, closingClient)
			close(closingClient)

		case msg := <-b.Notifier:
			for client := range b.clients {
				client <- msg
			}
		}
	}
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan []byte, 10)
	b.newClients <- messageChan
	defer func() {
		b.closingClients <- messageChan
	}()

	for {
		select {
		case <-r.Context().Done():
			return

		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}
