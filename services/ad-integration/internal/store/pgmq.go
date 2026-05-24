package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type QueueMessage struct {
	MessageID     string         `json:"message_id"`
	SchemaVersion string         `json:"schema_version"`
	EventType     string         `json:"event_type"`
	TenantID      string         `json:"tenant_id"`
	TraceID       string         `json:"trace_id"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Producer      string         `json:"producer"`
	Payload       map[string]any `json:"payload"`
}

type MessageQueue interface {
	Publish(queueName string, message QueueMessage) error
}

type PostgresQueue struct {
	db *sql.DB
}

func NewPostgresQueue(databaseURL string) (*PostgresQueue, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &PostgresQueue{db: db}, nil
}

func (q *PostgresQueue) Close() error {
	return q.db.Close()
}

func (q *PostgresQueue) Publish(queueName string, message QueueMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = q.db.ExecContext(context.Background(), `
		INSERT INTO ad_sync.pgmq_messages (queue_name, message_id, payload, status, created_at)
		VALUES ($1, $2, $3::jsonb, 'ready', now())
		ON CONFLICT (message_id) DO NOTHING
	`, queueName, message.MessageID, string(payload))
	return err
}
