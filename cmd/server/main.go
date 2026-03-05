package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/pgomes13/go-chat/internal/chat"
)

var (
	addr     = flag.String("addr", ":8080", "http service address")
	mongoURI = flag.String("mongo", "mongodb://localhost:27017", "mongodb URI")
)

func main() {
	flag.Parse()

	// Load .env file if present (errors are ignored — env vars may be set externally).
	godotenv.Load()

	// Cloud Run injects PORT; let it override the -addr flag.
	if port := os.Getenv("PORT"); port != "" {
		*addr = ":" + port
	}

	initAuth()
	s := initStore()

	hub := chat.NewHub(s)
	go hub.Run()

	registerRoutes(hub)

	log.Printf("server starting on %s", *addr)
	if err := http.ListenAndServe(*addr, secureHeaders(http.DefaultServeMux)); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
