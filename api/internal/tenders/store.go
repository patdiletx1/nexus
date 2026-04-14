package tenders

import (
	"sort"
	"sync"
	"time"
)

type Tender struct {
	ID            string         `json:"id,omitempty"`
	CompanyID     string         `json:"company_id"`
	ExternalID    string         `json:"external_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description,omitempty"`
	Region        string         `json:"region,omitempty"`
	ClosingAt     *time.Time     `json:"closing_at,omitempty"`
	PublishedAt   *time.Time     `json:"published_at,omitempty"`
	Source        string         `json:"source"`
	SourcePayload map[string]any `json:"source_payload,omitempty"`
	LastSyncedAt  time.Time      `json:"last_synced_at"`
	CreatedAt     time.Time      `json:"created_at,omitempty"`
	UpdatedAt     time.Time      `json:"updated_at,omitempty"`
}

type Store interface {
	Upsert(companyID string, tender Tender) bool
	ListByCompany(companyID string, limit int) []Tender
	GetByExternalID(companyID, externalID string) (Tender, bool)
}

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]Tender
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items: map[string]Tender{},
	}
}

func (s *InMemoryStore) Upsert(companyID string, tender Tender) bool {
	if companyID == "" || tender.ExternalID == "" || tender.Title == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := companyID + "::" + tender.ExternalID
	now := time.Now().UTC()
	current, exists := s.items[key]
	if exists {
		tender.ID = current.ID
		tender.CreatedAt = current.CreatedAt
		tender.UpdatedAt = now
	} else {
		tender.ID = ""
		tender.CreatedAt = now
		tender.UpdatedAt = now
	}
	tender.CompanyID = companyID
	tender.LastSyncedAt = now
	if tender.Source == "" {
		tender.Source = "chilecompra"
	}
	if tender.SourcePayload == nil {
		tender.SourcePayload = map[string]any{}
	}

	s.items[key] = tender
	return true
}

func (s *InMemoryStore) ListByCompany(companyID string, limit int) []Tender {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Tender, 0)
	for _, tender := range s.items {
		if tender.CompanyID == companyID {
			out = append(out, tender)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		left := out[i]
		right := out[j]
		if left.ClosingAt == nil && right.ClosingAt == nil {
			return left.UpdatedAt.After(right.UpdatedAt)
		}
		if left.ClosingAt == nil {
			return false
		}
		if right.ClosingAt == nil {
			return true
		}
		return left.ClosingAt.After(*right.ClosingAt)
	})

	if len(out) > limit {
		return out[:limit]
	}
	return out
}

func (s *InMemoryStore) GetByExternalID(companyID, externalID string) (Tender, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[companyID+"::"+externalID]
	if !ok {
		return Tender{}, false
	}
	return item, true
}
