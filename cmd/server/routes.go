package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/pgomes13/go-chat/internal/auth"
	"github.com/pgomes13/go-chat/internal/chat"
)

//go:embed static
var staticFiles embed.FS

func registerRoutes(hub *chat.Hub) {
	chat.SetAllowedOrigins(allowedOrigins())

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

// allowedOrigins builds the WebSocket origin allowlist.
// ALLOWED_ORIGINS (comma-separated) takes precedence; falls back to APP_BASE_URL.
func allowedOrigins() []string {
	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		var origins []string
		for _, o := range strings.Split(v, ",") {
			if s := strings.TrimSpace(o); s != "" {
				origins = append(origins, s)
			}
		}
		return origins
	}
	if base := os.Getenv("APP_BASE_URL"); base != "" {
		return []string{base}
	}
	return []string{"http://localhost" + *addr}
}

// secureHeaders sets standard HTTP security headers on every response.
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
