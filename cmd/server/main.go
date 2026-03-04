package main

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/pgomes13/go-chat/internal/auth"
	"github.com/pgomes13/go-chat/internal/chat"
	"github.com/pgomes13/go-chat/internal/store"
)

//go:embed static
var staticFiles embed.FS

var (
	addr      = flag.String("addr", ":8080", "http service address")
	redisAddr = flag.String("redis", "localhost:6379", "redis address")
)

func main() {
	flag.Parse()

	// Load .env file if present (errors are ignored — env vars may be set externally).
	godotenv.Load()

	// OAuth config from environment variables.
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost" + *addr + "/auth/google/callback"
	}
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = randomSecret()
		log.Println("WARNING: SESSION_SECRET not set — sessions will not persist across restarts")
	}
	auth.Init(clientID, clientSecret, redirectURL, sessionSecret)

	// Redis store (optional).
	var s *store.Store
	if rs, err := store.New(*redisAddr); err != nil {
		log.Printf("redis unavailable (%v) — running without persistence", err)
	} else {
		s = rs
	}

	hub := chat.NewHub(s)
	go hub.Run()

	// Auth routes.
	http.HandleFunc("/auth/google", auth.HandleLogin)
	http.HandleFunc("/auth/google/callback", auth.HandleCallback)
	http.HandleFunc("/auth/logout", auth.HandleLogout)
	http.HandleFunc("/me", auth.HandleMe)

	// WebSocket — requires authentication.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r)
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		chat.ServeWs(hub, w, r, user.Name)
	})

	// Static files.
	sub, _ := fs.Sub(staticFiles, "static")
	http.Handle("/", http.FileServer(http.FS(sub)))

	log.Printf("server starting on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func randomSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
