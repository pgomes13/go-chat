package main

import (
	"embed"
	"io/fs"
	"net/http"
	"net/url"
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
// Priority: ALLOWED_ORIGINS env var → APP_BASE_URL → origin extracted from OAUTH_REDIRECT_URL → localhost fallback.
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
	if redirect := os.Getenv("OAUTH_REDIRECT_URL"); redirect != "" {
		if u, err := url.Parse(redirect); err == nil {
			return []string{u.Scheme + "://" + u.Host}
		}
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
