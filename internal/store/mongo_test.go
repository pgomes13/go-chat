package store

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func mongoURI(t *testing.T) string {
	t.Helper()
	godotenv.Load("../../.env")
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	return uri
}

func TestConnection(t *testing.T) {
	s, err := New(mongoURI(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Log("connected successfully")
	s.client.Disconnect(context.Background())
}

func TestSaveAndHistory(t *testing.T) {
	s, err := New(mongoURI(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer s.client.Disconnect(context.Background())

	ctx := context.Background()

	msg, _ := json.Marshal(map[string]string{"sender": "test_user", "text": "hello from test"})
	if err := s.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	history, err := s.History(ctx)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(history) == 0 {
		t.Fatal("expected at least one message in history")
	}

	var last map[string]string
	json.Unmarshal(history[len(history)-1], &last)
	if last["text"] != "hello from test" {
		t.Errorf("last message text = %q, want %q", last["text"], "hello from test")
	}
	t.Logf("history has %d message(s); last: %v", len(history), last)
}
