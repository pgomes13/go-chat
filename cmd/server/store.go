package main

import (
	"log"
	"os"
	"strconv"

	"github.com/pgomes13/go-chat/internal/store"
)

func initStore() *store.Store {
	if envURI := os.Getenv("MONGO_URI"); envURI != "" {
		*mongoURI = envURI
	}

	db := envString("MONGO_DB", "gochat")
	historyLimit := envInt64("HISTORY_LIMIT", 50)

	s, err := store.New(*mongoURI, db, historyLimit)
	if err != nil {
		log.Printf("mongodb unavailable (%v) — running without persistence", err)
		return nil
	}
	return s
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
		log.Printf("invalid value for %s, using default %d", key, fallback)
	}
	return fallback
}
