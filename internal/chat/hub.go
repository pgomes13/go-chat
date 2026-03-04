package chat

import (
	"context"
	"log"

	"github.com/pgomes13/go-chat/internal/store"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients        map[*Client]bool
	broadcast      chan []byte // inbound from local clients
	localBroadcast chan []byte // inbound from Redis subscription
	register       chan *Client
	unregister     chan *Client
	store          *store.Store
}

func NewHub(s *store.Store) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		broadcast:      make(chan []byte),
		localBroadcast: make(chan []byte, 64),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		store:          s,
	}
}

func (h *Hub) Run() {
	ctx := context.Background()

	if h.store != nil {
		redisCh := h.store.Subscribe(ctx)
		go func() {
			for msg := range redisCh {
				h.localBroadcast <- msg
			}
		}()
	}

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
					log.Printf("redis save error: %v", err)
				}
				// Broadcast via Redis pub/sub; delivery comes back on localBroadcast.
				if err := h.store.Publish(ctx, message); err != nil {
					log.Printf("redis publish error: %v", err)
				}
			} else {
				h.deliver(message)
			}

		case message := <-h.localBroadcast:
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
