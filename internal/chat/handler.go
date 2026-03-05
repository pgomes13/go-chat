package chat

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var allowedOrigins []string

// SetAllowedOrigins configures which origins may open a WebSocket connection.
func SetAllowedOrigins(origins []string) {
	allowedOrigins = origins
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		for _, o := range allowedOrigins {
			if origin == o {
				return true
			}
		}
		return false
	},
}

// ServeWs handles websocket requests from clients.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID, username string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
	}
	client.hub.register <- client

	// Send message history to the new client.
	if hub.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		history, err := hub.store.History(ctx)
		if err != nil {
			log.Printf("mongodb history error: %v", err)
		}
		for _, msg := range history {
			client.send <- msg
		}
	}

	go client.writePump()
	go client.readPump()
}
