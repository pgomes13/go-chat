package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/pgomes13/go-chat/internal/auth"
	"github.com/pgomes13/go-chat/internal/chat"
)

//go:embed static
var staticFiles embed.FS

func registerRoutes(hub *chat.Hub) {
	http.HandleFunc("/auth/google", auth.HandleLogin)
	http.HandleFunc("/auth/google/callback", auth.HandleCallback)
	http.HandleFunc("/auth/logout", auth.HandleLogout)
	http.HandleFunc("/me", auth.HandleMe)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r)
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		chat.ServeWs(hub, w, r, user.ID, user.Name)
	})

	sub, _ := fs.Sub(staticFiles, "static")
	http.Handle("/", http.FileServer(http.FS(sub)))
}
