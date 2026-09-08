package audit

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	bufferSize    = 1024
	insertTimeout = 2 * time.Second
)

type Event struct {
	Timestamp   time.Time `bson:"ts" json:"ts"`
	Service     string    `bson:"service" json:"service"`
	Action      string    `bson:"action" json:"action"`
	ShortCode   string    `bson:"short_code,omitempty" json:"short_code,omitempty"`
	OriginalURL string    `bson:"original_url,omitempty" json:"original_url,omitempty"`
	Status      int       `bson:"status" json:"status"`
	LatencyMS   int64     `bson:"latency_ms" json:"latency_ms"`
	RequestID   string    `bson:"request_id" json:"request_id"`
	Method      string    `bson:"method" json:"method"`
	Path        string    `bson:"path" json:"path"`
	Error       string    `bson:"error,omitempty" json:"error,omitempty"`
}

type Emitter struct {
	client *mongo.Client
	col    *mongo.Collection
	events chan Event
	logger *slog.Logger
	wg     sync.WaitGroup
}

func Noop() *Emitter {
	return &Emitter{}
}

func Connect(ctx context.Context, uri, database, collection string, logger *slog.Logger) (*Emitter, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	col := client.Database(database).Collection(collection)
	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "ts", Value: -1}},
	})

	emitter := &Emitter{
		client: client,
		col:    col,
		events: make(chan Event, bufferSize),
		logger: logger,
	}
	emitter.wg.Add(1)
	go emitter.run()
	return emitter, nil
}

func (e *Emitter) Emit(ev Event) {
	if e == nil || e.events == nil {
		return
	}

	select {
	case e.events <- ev:
	default:
		if e.logger != nil {
			e.logger.Warn("audit event dropped", "action", ev.Action, "request_id", ev.RequestID)
		}
	}
}

func (e *Emitter) Close() {
	if e == nil || e.events == nil {
		return
	}

	close(e.events)
	e.wg.Wait()

	if e.client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = e.client.Disconnect(ctx)
}

func (e *Emitter) run() {
	defer e.wg.Done()

	for ev := range e.events {
		ctx, cancel := context.WithTimeout(context.Background(), insertTimeout)
		_, err := e.col.InsertOne(ctx, ev)
		cancel()
		if err != nil && e.logger != nil {
			e.logger.Error("failed to persist audit event", "err", err, "request_id", ev.RequestID)
		}
	}
}
