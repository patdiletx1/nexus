package idempotency

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresService struct {
	pool *pgxpool.Pool
}

func NewPostgresService(pool *pgxpool.Pool) *PostgresService {
	return &PostgresService{pool: pool}
}

func (s *PostgresService) Get(ctx context.Context, key Key) (StoredResponse, bool) {
	row := s.pool.QueryRow(ctx, `
		select status_code, response_payload
		from public.idempotency_keys
		where company_id = $1
		  and user_id = $2
		  and operation = $3
		  and resource_id = $4
		  and idempotency_key = $5
		  and expires_at > now()
		limit 1
	`,
		key.CompanyID,
		nullIfEmpty(key.UserID),
		key.Operation,
		key.ResourceID,
		key.IdempotencyKey,
	)

	var statusCode int
	var payloadRaw []byte
	if err := row.Scan(&statusCode, &payloadRaw); err != nil {
		return StoredResponse{}, false
	}

	payload := map[string]any{}
	if len(payloadRaw) > 0 {
		_ = json.Unmarshal(payloadRaw, &payload)
	}
	return StoredResponse{
		StatusCode: statusCode,
		Payload:    payload,
	}, true
}

func (s *PostgresService) Put(ctx context.Context, key Key, response StoredResponse, ttl time.Duration) {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if response.Payload == nil {
		response.Payload = map[string]any{}
	}

	_, err := s.pool.Exec(ctx, `
		insert into public.idempotency_keys (
			company_id, user_id, operation, resource_id, idempotency_key,
			status_code, response_payload, created_at, expires_at
		) values ($1, $2, $3, $4, $5, $6, $7, now(), now() + $8::interval)
		on conflict (company_id, user_id, operation, resource_id, idempotency_key) do nothing
	`,
		key.CompanyID,
		nullIfEmpty(key.UserID),
		key.Operation,
		key.ResourceID,
		key.IdempotencyKey,
		response.StatusCode,
		response.Payload,
		formatInterval(ttl),
	)
	if err != nil {
		log.Printf("idempotency_put_failed operation=%s resource_id=%s error=%v", key.Operation, key.ResourceID, err)
	}
}

func (s *PostgresService) CleanupExpired(ctx context.Context, limit int64) (int64, bool) {
	if limit <= 0 {
		limit = 500
	}

	tag, err := s.pool.Exec(ctx, `
		delete from public.idempotency_keys
		where id in (
			select id
			from public.idempotency_keys
			where expires_at <= now()
			order by expires_at asc
			limit $1
		)
	`, limit)
	if err != nil {
		log.Printf("idempotency_cleanup_failed error=%v", err)
		return 0, false
	}

	return tag.RowsAffected(), true
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func formatInterval(value time.Duration) string {
	seconds := int64(value.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10) + " seconds"
}
