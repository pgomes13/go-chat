package store

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	channelName    = "chat"
	historyKey     = "chat:history"
	maxHistorySize = 50
)

type Store struct {
	rdb *redis.Client
}

func New(addr string) (*Store, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	log.Printf("redis connected: %s", addr)
	return &Store{rdb: rdb}, nil
}

// SaveMessage appends a message to the history list, keeping the last maxHistorySize entries.
func (s *Store) SaveMessage(ctx context.Context, msg []byte) error {
	pipe := s.rdb.Pipeline()
	pipe.RPush(ctx, historyKey, msg)
	pipe.LTrim(ctx, historyKey, -maxHistorySize, -1)
	_, err := pipe.Exec(ctx)
	return err
}

// History returns all persisted messages in order.
func (s *Store) History(ctx context.Context) ([][]byte, error) {
	vals, err := s.rdb.LRange(ctx, historyKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	msgs := make([][]byte, len(vals))
	for i, v := range vals {
		msgs[i] = []byte(v)
	}
	return msgs, nil
}

// Publish sends a message to the Redis pub/sub channel.
func (s *Store) Publish(ctx context.Context, msg []byte) error {
	return s.rdb.Publish(ctx, channelName, msg).Err()
}

// Subscribe returns a channel that delivers messages from the Redis pub/sub channel.
func (s *Store) Subscribe(ctx context.Context) <-chan []byte {
	sub := s.rdb.Subscribe(ctx, channelName)
	ch := make(chan []byte, 64)
	go func() {
		defer close(ch)
		for msg := range sub.Channel() {
			ch <- []byte(msg.Payload)
		}
	}()
	return ch
}
