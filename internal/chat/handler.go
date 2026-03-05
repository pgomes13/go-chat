package chat

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for now
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
