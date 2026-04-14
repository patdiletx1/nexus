package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nexus/api/internal/chilecompra"
	"nexus/api/internal/companyprofile"
	"nexus/api/internal/tenders"
)

type fakeChileCompraClient struct {
	items []tenders.Tender
	err   error
}

func (f fakeChileCompraClient) FetchTenders(_ context.Context, _ *time.Time, _ int) ([]tenders.Tender, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func TestTendersSyncAndList(t *testing.T) {
	store := tenders.NewInMemoryStore()
	handler := TendersHandler{
		Store: store,
		Client: fakeChileCompraClient{
			items: []tenders.Tender{
				{
					ExternalID: "ext-1",
					Title:      "Licitacion Test",
					Source:     "chilecompra",
				},
			},
		},
	}

	syncReq := httptest.NewRequest(http.MethodGet, "/v1/tenders/sync?limit=10", nil)
	syncReq = syncReq.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	syncRec := httptest.NewRecorder()
	handler.Sync(syncRec, syncReq)
	if syncRec.Code != http.StatusOK {
		t.Fatalf("expected sync status %d, got %d", http.StatusOK, syncRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/tenders?limit=10", nil)
	listReq = listReq.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	listRec := httptest.NewRecorder()
	handler.List(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, listRec.Code)
	}
}

func TestTendersSyncNotConfigured(t *testing.T) {
	store := tenders.NewInMemoryStore()
	handler := TendersHandler{
		Store:  store,
		Client: chilecompra.NoopClient{},
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/tenders/sync", nil)
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Sync(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestTendersScore(t *testing.T) {
	store := tenders.NewInMemoryStore()
	store.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-score-1",
		Title:       "Servicio de Ingenieria en Terreno",
		Description: "Asistencia tecnica para obra vial",
		Region:      "Metropolitana",
		Source:      "chilecompra",
	})

	handler := TendersHandler{Store: store}
	req := httptest.NewRequest(http.MethodGet, "/v1/tenders/ext-score-1/score?company_region=Metropolitana&company_keywords=ingenieria,obra", nil)
	req.SetPathValue("id", "ext-score-1")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Score(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTendersScoreUsesStoredProfile(t *testing.T) {
	tendersStore := tenders.NewInMemoryStore()
	tendersStore.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-score-2",
		Title:       "Mantencion de Obras",
		Description: "Servicio de mantencion y obra",
		Region:      "Valparaiso",
		Source:      "chilecompra",
	})
	profileStore := companyprofile.NewInMemoryStore()
	profileStore.Upsert(companyprofile.Profile{
		CompanyID:       "company-1",
		PreferredRegion: "Valparaiso",
		Keywords:        []string{"mantencion", "obra"},
	})

	handler := TendersHandler{Store: tendersStore, Profile: profileStore}
	req := httptest.NewRequest(http.MethodGet, "/v1/tenders/ext-score-2/score", nil)
	req.SetPathValue("id", "ext-score-2")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Score(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTendersScoreCacheHit(t *testing.T) {
	tendersStore := tenders.NewInMemoryStore()
	tendersStore.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-score-cache",
		Title:       "Servicio de Mantencion",
		Description: "Mantencion electrica",
		Region:      "Biobio",
		Source:      "chilecompra",
	})

	handler := TendersHandler{
		Store:         tendersStore,
		ScoreCache:    tenders.NewInMemoryScoreCache(),
		ScoreCacheTTL: 15 * time.Minute,
	}

	makeReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/v1/tenders/ext-score-cache/score?company_region=Biobio&company_keywords=mantencion", nil)
		req.SetPathValue("id", "ext-score-cache")
		req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
		rec := httptest.NewRecorder()
		handler.Score(rec, req)
		return rec
	}

	first := makeReq()
	if first.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, first.Code)
	}
	second := makeReq()
	if second.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, second.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(second.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	cacheHit, ok := payload["cache_hit"].(bool)
	if !ok || !cacheHit {
		t.Fatalf("expected cache_hit=true, got %v", payload["cache_hit"])
	}
}

func TestTendersScoreWarmupWritesAndHits(t *testing.T) {
	tendersStore := tenders.NewInMemoryStore()
	tendersStore.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-warm-1",
		Title:       "Servicio de Mantencion",
		Description: "Mantencion electrica",
		Region:      "Biobio",
		Source:      "chilecompra",
	})
	tendersStore.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-warm-2",
		Title:       "Obra Vial",
		Description: "Ingenieria y obra",
		Region:      "Biobio",
		Source:      "chilecompra",
	})

	handler := TendersHandler{
		Store:         tendersStore,
		ScoreCache:    tenders.NewInMemoryScoreCache(),
		ScoreCacheTTL: 15 * time.Minute,
	}

	doWarmup := func() map[string]any {
		body := []byte(`{"limit":2,"company_region":"Biobio","company_keywords":["mantencion","obra"]}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/tenders/score/warmup", bytes.NewReader(body))
		req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
		rec := httptest.NewRecorder()
		handler.WarmupScoreCache(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed decoding payload: %v", err)
		}
		return payload
	}

	first := doWarmup()
	if int(first["cache_writes"].(float64)) != 2 {
		t.Fatalf("expected 2 cache writes, got %v", first["cache_writes"])
	}
	if int(first["cache_hits"].(float64)) != 0 {
		t.Fatalf("expected 0 cache hits in first run, got %v", first["cache_hits"])
	}

	second := doWarmup()
	if int(second["cache_writes"].(float64)) != 0 {
		t.Fatalf("expected 0 cache writes in second run, got %v", second["cache_writes"])
	}
	if int(second["cache_hits"].(float64)) != 2 {
		t.Fatalf("expected 2 cache hits in second run, got %v", second["cache_hits"])
	}
}

func TestTendersScoreWarmupUsesProfileFallback(t *testing.T) {
	tendersStore := tenders.NewInMemoryStore()
	tendersStore.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-warm-profile",
		Title:       "Servicio de Mantencion",
		Description: "Mantencion electrica",
		Region:      "Valparaiso",
		Source:      "chilecompra",
	})
	profileStore := companyprofile.NewInMemoryStore()
	profileStore.Upsert(companyprofile.Profile{
		CompanyID:       "company-1",
		PreferredRegion: "Valparaiso",
		Keywords:        []string{"mantencion"},
	})

	handler := TendersHandler{
		Store:         tendersStore,
		Profile:       profileStore,
		ScoreCache:    tenders.NewInMemoryScoreCache(),
		ScoreCacheTTL: 15 * time.Minute,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/tenders/score/warmup?limit=10", nil)
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.WarmupScoreCache(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	inputs, ok := payload["inputs"].(map[string]any)
	if !ok {
		t.Fatalf("expected inputs map, got %T", payload["inputs"])
	}
	if inputs["profile_source"] != "profile" {
		t.Fatalf("expected profile source profile, got %v", inputs["profile_source"])
	}
}

func TestTendersScoreWarmupTargetsSpecificIDs(t *testing.T) {
	tendersStore := tenders.NewInMemoryStore()
	tendersStore.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-warm-target-1",
		Title:       "Mantencion A",
		Description: "mantencion",
		Region:      "Biobio",
		Source:      "chilecompra",
	})
	tendersStore.Upsert("company-1", tenders.Tender{
		ExternalID:  "ext-warm-target-2",
		Title:       "Mantencion B",
		Description: "mantencion",
		Region:      "Biobio",
		Source:      "chilecompra",
	})

	handler := TendersHandler{
		Store:         tendersStore,
		ScoreCache:    tenders.NewInMemoryScoreCache(),
		ScoreCacheTTL: 15 * time.Minute,
	}

	body := []byte(`{"tender_ids":["ext-warm-target-1","ext-warm-target-1","missing-id"],"company_region":"Biobio","company_keywords":["mantencion"]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenders/score/warmup", bytes.NewReader(body))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.WarmupScoreCache(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	processedCount := int(payload["processed_count"].(float64))
	if processedCount != 1 {
		t.Fatalf("expected processed_count=1, got %d", processedCount)
	}
	cacheWrites := int(payload["cache_writes"].(float64))
	if cacheWrites != 1 {
		t.Fatalf("expected cache_writes=1, got %d", cacheWrites)
	}
}
