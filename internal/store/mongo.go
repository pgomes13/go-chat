package store

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/pgomes13/go-chat/internal/commons"
)

type Store struct {
	client       *mongo.Client
	coll         *mongo.Collection
	historyLimit int64
}

func New(uri, db string, historyLimit int64) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	coll := client.Database(db).Collection(commons.CollName)
	coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: 1}},
	})

	log.Printf("mongodb connected: %s", redactURI(uri))
	return &Store{client: client, coll: coll, historyLimit: historyLimit}, nil
}

type messageDoc struct {
	SenderID  string    `bson:"sender_id"`
	Sender    string    `bson:"sender"`
	Text      string    `bson:"text"`
	CreatedAt time.Time `bson:"created_at"`
}

// SaveMessage parses the JSON payload and inserts it as a document.
func (s *Store) SaveMessage(ctx context.Context, msg []byte) error {
	var p struct {
		SenderID string `json:"sender_id"`
		Sender   string `json:"sender"`
		Text     string `json:"text"`
	}
	if err := json.Unmarshal(msg, &p); err != nil {
		return err
	}
	_, err := s.coll.InsertOne(ctx, messageDoc{
		SenderID:  p.SenderID,
		Sender:    p.Sender,
		Text:      p.Text,
		CreatedAt: time.Now().UTC(),
	})
	return err
}

// History returns the last maxHistorySize messages in chronological order.
func (s *Store) History(ctx context.Context) ([][]byte, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(s.historyLimit)

	cursor, err := s.coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []messageDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	// Reverse to chronological order.
	for i, j := 0, len(docs)-1; i < j; i, j = i+1, j-1 {
		docs[i], docs[j] = docs[j], docs[i]
	}

	msgs := make([][]byte, 0, len(docs))
	for _, d := range docs {
		b, _ := json.Marshal(map[string]string{
			"sender_id": d.SenderID,
			"sender":    d.Sender,
			"text":      d.Text,
		})
		msgs = append(msgs, b)
	}
	return msgs, nil
}

func redactURI(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.User == nil {
		return uri
	}
	u.User = url.User(u.User.Username())
	return u.String()
}
