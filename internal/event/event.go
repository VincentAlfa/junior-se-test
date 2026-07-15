package event

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRetrying  Status = "retrying"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

type Event struct {
	ID            string          `json:"event_id"`
	CustomerID    string          `json:"customer_id"`
	Payload       json.RawMessage `json:"payload"`
	Status        Status          `json:"status"`
	Attempts      int             `json:"attempts"`
	CreatedAt     time.Time       `json:"created_at"`
	LastAttemptAt time.Time       `json:"last_attempt_at,omitempty"`
	NextRetryAt   time.Time       `json:"next_retry_at,omitempty"`
}
