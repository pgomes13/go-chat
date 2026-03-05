package main

import (
	"log"
	"os"

	"github.com/pgomes13/go-chat/internal/store"
)

func initStore() *store.Store {
	if envURI := os.Getenv("MONGO_URI"); envURI != "" {
		*mongoURI = envURI
	}

	s, err := store.New(*mongoURI)
	if err != nil {
		log.Printf("mongodb unavailable (%v) — running without persistence", err)
		return nil
	}
	return s
}
