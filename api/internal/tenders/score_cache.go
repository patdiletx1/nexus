package tenders

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type CachedScore struct {
	Score   int
	Reasons []string
}

type ScoreCache interface {
	Get(companyID, externalID, profileFingerprint string) (CachedScore, bool)
	Put(companyID, externalID, profileFingerprint string, score CachedScore, ttl time.Duration)
	InvalidateCompany(companyID string) (int64, bool)
}

type NoopScoreCache struct{}

func (NoopScoreCache) Get(_, _, _ string) (CachedScore, bool) {
	return CachedScore{}, false
}

func (NoopScoreCache) Put(_, _, _ string, _ CachedScore, _ time.Duration) {}

func (NoopScoreCache) InvalidateCompany(_ string) (int64, bool) {
	return 0, false
}

type inMemoryCachedScore struct {
	score     CachedScore
	expiresAt time.Time
}

type InMemoryScoreCache struct {
	mu    sync.RWMutex
	items map[string]inMemoryCachedScore
}

func NewInMemoryScoreCache() *InMemoryScoreCache {
	return &InMemoryScoreCache{
		items: map[string]inMemoryCachedScore{},
	}
}

func (c *InMemoryScoreCache) Get(companyID, externalID, profileFingerprint string) (CachedScore, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := scoreCacheKey(companyID, externalID, profileFingerprint)
	entry, ok := c.items[key]
	if !ok {
		return CachedScore{}, false
	}
	if entry.expiresAt.Before(time.Now().UTC()) {
		return CachedScore{}, false
	}
	return entry.score, true
}

func (c *InMemoryScoreCache) Put(companyID, externalID, profileFingerprint string, score CachedScore, ttl time.Duration) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[scoreCacheKey(companyID, externalID, profileFingerprint)] = inMemoryCachedScore{
		score:     score,
		expiresAt: time.Now().UTC().Add(ttl),
	}
}

func (c *InMemoryScoreCache) InvalidateCompany(companyID string) (int64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := companyID + "::"
	deleted := int64(0)
	for key := range c.items {
		if strings.HasPrefix(key, prefix) {
			delete(c.items, key)
			deleted++
		}
	}
	return deleted, true
}

func scoreCacheKey(companyID, externalID, profileFingerprint string) string {
	return companyID + "::" + externalID + "::" + profileFingerprint
}

func BuildProfileFingerprint(region string, keywords []string) string {
	normalizedRegion := strings.ToLower(strings.TrimSpace(region))

	normalizedKeywords := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		trimmed := strings.ToLower(strings.TrimSpace(keyword))
		if trimmed != "" {
			normalizedKeywords = append(normalizedKeywords, trimmed)
		}
	}
	sort.Strings(normalizedKeywords)

	return normalizedRegion + "|" + strings.Join(normalizedKeywords, ",")
}
