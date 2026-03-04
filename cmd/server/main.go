package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"

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

	var s *store.Store
	if rs, err := store.New(*redisAddr); err != nil {
		log.Printf("redis unavailable (%v) — running without persistence", err)
	} else {
		s = rs
	}

	hub := chat.NewHub(s)
	go hub.Run()

	sub, _ := fs.Sub(staticFiles, "static")
	http.Handle("/", http.FileServer(http.FS(sub)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWs(hub, w, r)
	})

	log.Printf("server starting on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
