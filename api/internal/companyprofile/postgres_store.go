package companyprofile

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Get(companyID string) (Profile, bool) {
	row := s.pool.QueryRow(context.Background(), `
		select company_id, preferred_region, keywords, updated_by_user_id, created_at, updated_at
		from public.company_scoring_profiles
		where company_id = $1
		limit 1
	`, companyID)

	var profile Profile
	var preferredRegion *string
	var updatedByUserID *string
	if err := row.Scan(
		&profile.CompanyID,
		&preferredRegion,
		&profile.Keywords,
		&updatedByUserID,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return Profile{}, false
	}
	if preferredRegion != nil {
		profile.PreferredRegion = *preferredRegion
	}
	if updatedByUserID != nil {
		profile.UpdatedByUserID = *updatedByUserID
	}
	if profile.Keywords == nil {
		profile.Keywords = []string{}
	}
	return profile, true
}

func (s *PostgresStore) Upsert(profile Profile) bool {
	if profile.CompanyID == "" {
		return false
	}
	if profile.Keywords == nil {
		profile.Keywords = []string{}
	}

	_, err := s.pool.Exec(context.Background(), `
		insert into public.company_scoring_profiles (
			company_id, preferred_region, keywords, updated_by_user_id, created_at, updated_at
		) values ($1, $2, $3, $4, now(), now())
		on conflict (company_id) do update
		  set preferred_region = excluded.preferred_region,
		      keywords = excluded.keywords,
		      updated_by_user_id = excluded.updated_by_user_id,
		      updated_at = now()
	`,
		profile.CompanyID,
		nullIfEmpty(profile.PreferredRegion),
		profile.Keywords,
		nullIfEmpty(profile.UpdatedByUserID),
	)
	if err != nil {
		log.Printf("company_profile_upsert_failed company_id=%s error=%v", profile.CompanyID, err)
		return false
	}
	return true
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
