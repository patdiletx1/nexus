package tenders

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresScoreCache struct {
	pool *pgxpool.Pool
}

func NewPostgresScoreCache(pool *pgxpool.Pool) *PostgresScoreCache {
	return &PostgresScoreCache{pool: pool}
}

func (c *PostgresScoreCache) Get(companyID, externalID, profileFingerprint string) (CachedScore, bool) {
	row := c.pool.QueryRow(context.Background(), `
		select score, reasons
		from public.tender_score_cache
		where company_id = $1
		  and tender_external_id = $2
		  and profile_fingerprint = $3
		  and expires_at > now()
		limit 1
	`, companyID, externalID, profileFingerprint)

	var score int
	var reasons []string
	if err := row.Scan(&score, &reasons); err != nil {
		return CachedScore{}, false
	}
	if reasons == nil {
		reasons = []string{}
	}
	return CachedScore{
		Score:   score,
		Reasons: reasons,
	}, true
}

func (c *PostgresScoreCache) Put(companyID, externalID, profileFingerprint string, score CachedScore, ttl time.Duration) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	if score.Reasons == nil {
		score.Reasons = []string{}
	}

	_, err := c.pool.Exec(context.Background(), `
		insert into public.tender_score_cache (
			company_id, tender_external_id, profile_fingerprint, score, reasons, expires_at, created_at, updated_at
		) values ($1, $2, $3, $4, $5, now() + $6::interval, now(), now())
		on conflict (company_id, tender_external_id, profile_fingerprint) do update
		  set score = excluded.score,
		      reasons = excluded.reasons,
		      expires_at = excluded.expires_at,
		      updated_at = now()
	`,
		companyID,
		externalID,
		profileFingerprint,
		score.Score,
		score.Reasons,
		formatInterval(ttl),
	)
	if err != nil {
		log.Printf("tender_score_cache_put_failed company_id=%s tender=%s error=%v", companyID, externalID, err)
	}
}

func (c *PostgresScoreCache) InvalidateCompany(companyID string) (int64, bool) {
	tag, err := c.pool.Exec(context.Background(), `
		delete from public.tender_score_cache
		where company_id = $1
	`, companyID)
	if err != nil {
		log.Printf("tender_score_cache_invalidate_failed company_id=%s error=%v", companyID, err)
		return 0, false
	}
	return tag.RowsAffected(), true
}

func formatInterval(value time.Duration) string {
	seconds := int64(value.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10) + " seconds"
}
