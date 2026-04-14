package companyprofile

import (
	"sync"
	"time"
)

type Profile struct {
	CompanyID       string    `json:"company_id"`
	PreferredRegion string    `json:"preferred_region,omitempty"`
	Keywords        []string  `json:"keywords"`
	UpdatedByUserID string    `json:"updated_by_user_id,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type Store interface {
	Get(companyID string) (Profile, bool)
	Upsert(profile Profile) bool
}

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]Profile
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items: map[string]Profile{},
	}
}

func (s *InMemoryStore) Get(companyID string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[companyID]
	return item, ok
}

func (s *InMemoryStore) Upsert(profile Profile) bool {
	if profile.CompanyID == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	current, exists := s.items[profile.CompanyID]
	if exists {
		profile.CreatedAt = current.CreatedAt
		profile.UpdatedAt = now
	} else {
		profile.CreatedAt = now
		profile.UpdatedAt = now
	}
	if profile.Keywords == nil {
		profile.Keywords = []string{}
	}

	s.items[profile.CompanyID] = profile
	return true
}
