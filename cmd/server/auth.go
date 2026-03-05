package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"

	"github.com/pgomes13/go-chat/internal/auth"
)

func initAuth() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	if redirectURL == "" {
		base := os.Getenv("APP_BASE_URL")
		if base == "" {
			base = "http://localhost" + *addr
		}
		redirectURL = base + "/auth/google/callback"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = randomSecret()
		log.Println("WARNING: SESSION_SECRET not set — sessions will not persist across restarts")
	}

	auth.Init(clientID, clientSecret, redirectURL, sessionSecret)
}

func randomSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
