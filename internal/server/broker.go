package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Broker struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{} // map of active clients
	ctx     context.Context
}

func NewBroker(ctx context.Context) *Broker {
	b := &Broker{
		clients: make(map[chan []byte]struct{}),
		ctx:     ctx,
	}

	return b
}

func (b *Broker) Notify(msg []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		select {
		case client <- msg:
		default:
		}
	}
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan []byte, 1)

	b.mu.Lock()
	b.clients[messageChan] = struct{}{}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, messageChan)
		b.mu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Flusher unsupported", http.StatusInternalServerError)
		log.Println("Flusher unsupported")
		return
	}

	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return

		case <-b.ctx.Done():
			return

		case msg := <-messageChan:
			_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
			if err != nil {
				return
			}
			flusher.Flush()

		}
	}
}
