package store

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	godotenv.Load("../../.env")

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	db := os.Getenv("MONGO_DB")
	if db == "" {
		db = "gochat"
	}

	s, err := New(uri, db, 50)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return s
}

func TestConnection(t *testing.T) {
	s := testStore(t)
	t.Log("connected successfully")
	s.client.Disconnect(context.Background())
}

func TestSaveAndHistory(t *testing.T) {
	s := testStore(t)
	defer s.client.Disconnect(context.Background())

	ctx := context.Background()

	msg, _ := json.Marshal(map[string]string{
		"sender_id": "test_id",
		"sender":    "test_user",
		"text":      "hello from test",
	})
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
