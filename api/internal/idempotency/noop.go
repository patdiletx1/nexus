package idempotency

import (
	"context"
	"time"
)

type NoopService struct{}

func (NoopService) Get(_ context.Context, _ Key) (StoredResponse, bool) {
	return StoredResponse{}, false
}

func (NoopService) Put(_ context.Context, _ Key, _ StoredResponse, _ time.Duration) {}

func (NoopService) CleanupExpired(_ context.Context, _ int64) (int64, bool) {
	return 0, false
}
