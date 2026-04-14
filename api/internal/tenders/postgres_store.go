package tenders

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Upsert(companyID string, tender Tender) bool {
	if companyID == "" || tender.ExternalID == "" || tender.Title == "" {
		return false
	}

	now := time.Now().UTC()
	if tender.Source == "" {
		tender.Source = "chilecompra"
	}
	if tender.SourcePayload == nil {
		tender.SourcePayload = map[string]any{}
	}

	_, err := s.pool.Exec(context.Background(), `
		insert into public.tenders (
			company_id, external_id, title, description, region, closing_at, published_at,
			source, source_payload, last_synced_at, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
		on conflict (company_id, external_id) do update
		  set title = excluded.title,
		      description = excluded.description,
		      region = excluded.region,
		      closing_at = excluded.closing_at,
		      published_at = excluded.published_at,
		      source = excluded.source,
		      source_payload = excluded.source_payload,
		      last_synced_at = excluded.last_synced_at,
		      updated_at = now()
	`,
		companyID,
		tender.ExternalID,
		tender.Title,
		nullIfEmpty(tender.Description),
		nullIfEmpty(tender.Region),
		tender.ClosingAt,
		tender.PublishedAt,
		tender.Source,
		tender.SourcePayload,
		now,
	)
	if err != nil {
		log.Printf("tenders_upsert_failed company_id=%s external_id=%s error=%v", companyID, tender.ExternalID, err)
		return false
	}
	return true
}

func (s *PostgresStore) ListByCompany(companyID string, limit int) []Tender {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := s.pool.Query(context.Background(), `
		select id, company_id, external_id, title, description, region, closing_at, published_at,
		       source, source_payload, last_synced_at, created_at, updated_at
		from public.tenders
		where company_id = $1
		order by closing_at desc nulls last, updated_at desc
		limit $2
	`, companyID, limit)
	if err != nil {
		log.Printf("tenders_list_failed company_id=%s error=%v", companyID, err)
		return []Tender{}
	}
	defer rows.Close()

	items := make([]Tender, 0)
	for rows.Next() {
		var item Tender
		var description *string
		var region *string
		if err := rows.Scan(
			&item.ID,
			&item.CompanyID,
			&item.ExternalID,
			&item.Title,
			&description,
			&region,
			&item.ClosingAt,
			&item.PublishedAt,
			&item.Source,
			&item.SourcePayload,
			&item.LastSyncedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			log.Printf("tenders_list_scan_failed company_id=%s error=%v", companyID, err)
			continue
		}
		if description != nil {
			item.Description = *description
		}
		if region != nil {
			item.Region = *region
		}
		if item.SourcePayload == nil {
			item.SourcePayload = map[string]any{}
		}
		items = append(items, item)
	}
	return items
}

func (s *PostgresStore) GetByExternalID(companyID, externalID string) (Tender, bool) {
	row := s.pool.QueryRow(context.Background(), `
		select id, company_id, external_id, title, description, region, closing_at, published_at,
		       source, source_payload, last_synced_at, created_at, updated_at
		from public.tenders
		where company_id = $1 and external_id = $2
		limit 1
	`, companyID, externalID)

	var item Tender
	var description *string
	var region *string
	if err := row.Scan(
		&item.ID,
		&item.CompanyID,
		&item.ExternalID,
		&item.Title,
		&description,
		&region,
		&item.ClosingAt,
		&item.PublishedAt,
		&item.Source,
		&item.SourcePayload,
		&item.LastSyncedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Tender{}, false
	}
	if description != nil {
		item.Description = *description
	}
	if region != nil {
		item.Region = *region
	}
	if item.SourcePayload == nil {
		item.SourcePayload = map[string]any{}
	}
	return item, true
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
