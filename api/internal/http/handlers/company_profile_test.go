package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nexus/api/internal/companyprofile"
	"nexus/api/internal/tenders"
)

type fakeScoreCache struct {
	invalidatedCompanyID string
}

func (f *fakeScoreCache) Get(_, _, _ string) (tenders.CachedScore, bool) {
	return tenders.CachedScore{}, false
}

func (f *fakeScoreCache) Put(_, _, _ string, _ tenders.CachedScore, _ time.Duration) {}

func (f *fakeScoreCache) InvalidateCompany(companyID string) (int64, bool) {
	f.invalidatedCompanyID = companyID
	return 1, true
}

func TestCompanyProfileUpsertInvalidatesScoreCache(t *testing.T) {
	store := companyprofile.NewInMemoryStore()
	cache := &fakeScoreCache{}
	handler := CompanyProfileHandler{
		Store:      store,
		ScoreCache: cache,
	}

	req := httptest.NewRequest(http.MethodPut, "/v1/company/profile", bytes.NewReader([]byte(`{"preferred_region":"Metropolitana","keywords":["ingenieria","obra"]}`)))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Upsert(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if cache.invalidatedCompanyID != "company-1" {
		t.Fatalf("expected invalidation for company-1, got %s", cache.invalidatedCompanyID)
	}
}
