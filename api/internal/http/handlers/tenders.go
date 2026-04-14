package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nexus/api/internal/chilecompra"
	"nexus/api/internal/companyprofile"
	"nexus/api/internal/tenders"
)

type TendersHandler struct {
	Store         tenders.Store
	Client        chilecompra.Client
	Profile       companyprofile.Store
	ScoreCache    tenders.ScoreCache
	ScoreCacheTTL time.Duration
}

type scoreWarmupRequest struct {
	Limit           int      `json:"limit"`
	CompanyRegion   string   `json:"company_region"`
	CompanyKeywords []string `json:"company_keywords"`
	TenderIDs       []string `json:"tender_ids"`
}

func (h TendersHandler) Sync(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil || h.Client == nil {
		http.Error(w, "tenders service unavailable", http.StatusInternalServerError)
		return
	}

	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"), 50, 200)
	since, err := parseOptionalTime(r.URL.Query().Get("since"))
	if err != nil {
		http.Error(w, "invalid since value, expected RFC3339", http.StatusBadRequest)
		return
	}

	items, err := h.Client.FetchTenders(r.Context(), since, limit)
	if err == chilecompra.ErrNotConfigured {
		http.Error(w, "chilecompra client not configured", http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		http.Error(w, "failed to sync tenders", http.StatusBadGateway)
		return
	}

	upserted := 0
	for _, item := range items {
		if h.Store.Upsert(companyID, item) {
			upserted++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"synced_count":   len(items),
		"upserted_count": upserted,
		"limit":          limit,
		"since":          formatTimePointer(since),
	})
}

func (h TendersHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		http.Error(w, "tenders store unavailable", http.StatusInternalServerError)
		return
	}
	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"), 50, 200)
	items := h.Store.ListByCompany(companyID, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"tenders":        items,
		"returned_count": len(items),
	})
}

func (h TendersHandler) Score(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		http.Error(w, "tenders store unavailable", http.StatusInternalServerError)
		return
	}
	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	externalID := strings.TrimSpace(r.PathValue("id"))
	if externalID == "" {
		http.Error(w, "tender id is required", http.StatusBadRequest)
		return
	}

	item, ok := h.Store.GetByExternalID(companyID, externalID)
	if !ok {
		http.Error(w, "tender not found", http.StatusNotFound)
		return
	}

	companyRegion, keywords, profileSource := h.resolveScoreInputs(
		companyID,
		strings.TrimSpace(r.URL.Query().Get("company_region")),
		parseKeywords(r.URL.Query().Get("company_keywords")),
	)
	result := tenders.ScoreTender(item, tenders.ScoreInput{
		CompanyRegion:   companyRegion,
		CompanyKeywords: keywords,
	})
	profileFingerprint := tenders.BuildProfileFingerprint(companyRegion, keywords)
	cacheHit := false
	if h.ScoreCache != nil {
		if cached, ok := h.ScoreCache.Get(companyID, item.ExternalID, profileFingerprint); ok {
			result.Score = cached.Score
			result.Reasons = cached.Reasons
			cacheHit = true
		} else {
			h.ScoreCache.Put(companyID, item.ExternalID, profileFingerprint, tenders.CachedScore{
				Score:   result.Score,
				Reasons: result.Reasons,
			}, h.ScoreCacheTTL)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tender_id": item.ExternalID,
		"score":     result.Score,
		"reasons":   result.Reasons,
		"cache_hit": cacheHit,
		"inputs": map[string]any{
			"company_region":   companyRegion,
			"company_keywords": keywords,
			"profile_source":   profileSource,
		},
	})
}

func (h TendersHandler) WarmupScoreCache(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		http.Error(w, "tenders store unavailable", http.StatusInternalServerError)
		return
	}
	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	var req scoreWarmupRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	limit := req.Limit
	if limit <= 0 {
		limit = parseLimit(r.URL.Query().Get("limit"), 100, 500)
	}
	if limit <= 0 {
		limit = 100
	}

	profileKeywords := req.CompanyKeywords
	if len(profileKeywords) == 0 {
		profileKeywords = parseKeywords(r.URL.Query().Get("company_keywords"))
	}
	tenderIDs := req.TenderIDs
	if len(tenderIDs) == 0 {
		tenderIDs = parseKeywords(r.URL.Query().Get("tender_ids"))
	}
	companyRegion, keywords, profileSource := h.resolveScoreInputs(
		companyID,
		strings.TrimSpace(req.CompanyRegion),
		profileKeywords,
	)
	fingerprint := tenders.BuildProfileFingerprint(companyRegion, keywords)
	items := h.resolveWarmupTenders(companyID, limit, tenderIDs)
	if h.ScoreCache == nil {
		h.ScoreCache = tenders.NoopScoreCache{}
	}

	cacheHits := 0
	cachedWrites := 0
	processedCount := 0
	for _, item := range items {
		processedCount++
		if _, ok := h.ScoreCache.Get(companyID, item.ExternalID, fingerprint); ok {
			cacheHits++
			continue
		}
		score := tenders.ScoreTender(item, tenders.ScoreInput{
			CompanyRegion:   companyRegion,
			CompanyKeywords: keywords,
		})
		h.ScoreCache.Put(companyID, item.ExternalID, fingerprint, tenders.CachedScore{
			Score:   score.Score,
			Reasons: score.Reasons,
		}, h.ScoreCacheTTL)
		cachedWrites++
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"processed_count": processedCount,
		"cache_hits":      cacheHits,
		"cache_writes":    cachedWrites,
		"limit":           limit,
		"targeted_ids":    tenderIDs,
		"inputs": map[string]any{
			"company_region":   companyRegion,
			"company_keywords": keywords,
			"profile_source":   profileSource,
		},
	})
}

func (h TendersHandler) resolveWarmupTenders(companyID string, limit int, tenderIDs []string) []tenders.Tender {
	if len(tenderIDs) == 0 {
		return h.Store.ListByCompany(companyID, limit)
	}
	seen := map[string]struct{}{}
	out := make([]tenders.Tender, 0, len(tenderIDs))
	for _, rawID := range tenderIDs {
		externalID := strings.TrimSpace(rawID)
		if externalID == "" {
			continue
		}
		if _, exists := seen[externalID]; exists {
			continue
		}
		seen[externalID] = struct{}{}
		if item, ok := h.Store.GetByExternalID(companyID, externalID); ok {
			out = append(out, item)
		}
	}
	return out
}

func (h TendersHandler) resolveScoreInputs(companyID, companyRegion string, keywords []string) (string, []string, string) {
	profileSource := "query"
	if companyRegion == "" || len(keywords) == 0 {
		if h.Profile != nil {
			if profile, ok := h.Profile.Get(companyID); ok {
				if companyRegion == "" {
					companyRegion = profile.PreferredRegion
				}
				if len(keywords) == 0 {
					keywords = profile.Keywords
				}
				profileSource = "profile"
			} else {
				profileSource = "default"
			}
		} else {
			profileSource = "default"
		}
	}
	return companyRegion, keywords, profileSource
}

func parseLimit(raw string, fallback, max int) int {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	if parsed > max {
		return max
	}
	return parsed
}

func formatTimePointer(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func parseKeywords(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
