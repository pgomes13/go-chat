package chat

import (
	"context"
	"log"

	"github.com/pgomes13/go-chat/internal/store"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	store      *store.Store
}

func NewHub(s *store.Store) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		store:      s,
	}
}

func (h *Hub) Run() {
	ctx := context.Background()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			if h.store != nil {
				if err := h.store.SaveMessage(ctx, message); err != nil {
					log.Printf("mongodb save error: %v", err)
				}
			}
			h.deliver(message)
		}
	}
}

func (h *Hub) deliver(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}
