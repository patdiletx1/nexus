package audit

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresLogger struct {
	pool *pgxpool.Pool
}

func NewPostgresLogger(pool *pgxpool.Pool) *PostgresLogger {
	return &PostgresLogger{pool: pool}
}

func (l *PostgresLogger) LogEvent(event Event) {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.Payload == nil {
		event.Payload = map[string]any{}
	}

	_, err := l.pool.Exec(context.Background(), `
		insert into public.audit_events (
			company_id, actor_user_id, event_type, entity_type, entity_id, payload, created_at
		) values ($1, $2, $3, $4, $5, $6, $7)
	`,
		event.CompanyID,
		nullIfEmpty(event.ActorUserID),
		event.EventType,
		event.EntityType,
		nullIfEmpty(event.EntityID),
		event.Payload,
		event.CreatedAt,
	)
	if err != nil {
		log.Printf("audit_log_event_failed event_type=%s company_id=%s error=%v", event.EventType, event.CompanyID, err)
	}
}

func (l *PostgresLogger) ListEvents(query EventQuery) []Event {
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}

	conditions, args := buildEventFilters(query)

	args = append(args, query.Limit)
	sql := fmt.Sprintf(`
		select id, company_id, actor_user_id, event_type, entity_type, entity_id, payload, created_at
		from public.audit_events
		where %s
		order by created_at desc, id desc
		limit $%d
	`, strings.Join(conditions, " and "), len(args))

	rows, err := l.pool.Query(context.Background(), sql, args...)
	if err != nil {
		log.Printf("audit_list_events_failed company_id=%s entity_type=%s entity_id=%s error=%v", query.CompanyID, query.EntityType, query.EntityID, err)
		return []Event{}
	}
	defer rows.Close()

	events := make([]Event, 0)
	for rows.Next() {
		var event Event
		var actorUserID *string
		if err := rows.Scan(
			&event.ID,
			&event.CompanyID,
			&actorUserID,
			&event.EventType,
			&event.EntityType,
			&event.EntityID,
			&event.Payload,
			&event.CreatedAt,
		); err != nil {
			log.Printf("audit_list_events_scan_failed company_id=%s entity_id=%s error=%v", query.CompanyID, query.EntityID, err)
			continue
		}
		if actorUserID != nil {
			event.ActorUserID = *actorUserID
		}
		if event.Payload == nil {
			event.Payload = map[string]any{}
		}
		events = append(events, event)
	}
	return events
}

func (l *PostgresLogger) CountEvents(query EventQuery) (int64, bool) {
	conditions, args := buildEventFilters(query)
	sql := fmt.Sprintf(`
		select count(*)::bigint
		from public.audit_events
		where %s
	`, strings.Join(conditions, " and "))

	var total int64
	if err := l.pool.QueryRow(context.Background(), sql, args...).Scan(&total); err != nil {
		log.Printf("audit_count_events_failed company_id=%s entity_type=%s entity_id=%s error=%v", query.CompanyID, query.EntityType, query.EntityID, err)
		return 0, false
	}
	return total, true
}

func buildEventFilters(query EventQuery) ([]string, []any) {
	conditions := []string{
		"company_id = $1",
		"entity_type = $2",
		"entity_id = $3",
	}
	args := []any{query.CompanyID, query.EntityType, query.EntityID}

	if query.EventType != "" {
		args = append(args, query.EventType)
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", len(args)))
	}
	if query.From != nil {
		args = append(args, *query.From)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if query.To != nil {
		args = append(args, *query.To)
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)))
	}
	if query.BeforeCreatedAt != nil && query.BeforeEventID != "" {
		args = append(args, *query.BeforeCreatedAt, query.BeforeEventID)
		conditions = append(conditions, fmt.Sprintf("(created_at, id) < ($%d, $%d::uuid)", len(args)-1, len(args)))
	}

	return conditions, args
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
