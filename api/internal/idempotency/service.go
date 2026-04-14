package idempotency

import (
	"context"
	"time"
)

type Key struct {
	CompanyID      string
	UserID         string
	Operation      string
	ResourceID     string
	IdempotencyKey string
}

type StoredResponse struct {
	StatusCode int
	Payload    map[string]any
}

type Service interface {
	Get(ctx context.Context, key Key) (StoredResponse, bool)
	Put(ctx context.Context, key Key, response StoredResponse, ttl time.Duration)
	CleanupExpired(ctx context.Context, limit int64) (int64, bool)
}
