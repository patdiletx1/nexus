package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"nexus/api/internal/companyprofile"
	"nexus/api/internal/tenders"
)

type CompanyProfileHandler struct {
	Store      companyprofile.Store
	ScoreCache tenders.ScoreCache
}

type upsertProfileRequest struct {
	PreferredRegion string   `json:"preferred_region"`
	Keywords        []string `json:"keywords"`
}

func (h CompanyProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		http.Error(w, "profile store unavailable", http.StatusInternalServerError)
		return
	}
	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	profile, ok := h.Store.Get(companyID)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"profile": map[string]any{
				"company_id":       companyID,
				"preferred_region": "",
				"keywords":         []string{},
			},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": profile})
}

func (h CompanyProfileHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		http.Error(w, "profile store unavailable", http.StatusInternalServerError)
		return
	}
	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}
	userID := UserIDFromContext(r.Context())

	var req upsertProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	normalizedKeywords := make([]string, 0, len(req.Keywords))
	seen := map[string]struct{}{}
	for _, keyword := range req.Keywords {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalizedKeywords = append(normalizedKeywords, trimmed)
	}

	profile := companyprofile.Profile{
		CompanyID:       companyID,
		PreferredRegion: strings.TrimSpace(req.PreferredRegion),
		Keywords:        normalizedKeywords,
		UpdatedByUserID: userID,
	}
	if !h.Store.Upsert(profile) {
		http.Error(w, "failed to save profile", http.StatusInternalServerError)
		return
	}
	if h.ScoreCache != nil {
		h.ScoreCache.InvalidateCompany(companyID)
	}
	profile, _ = h.Store.Get(companyID)
	writeJSON(w, http.StatusOK, map[string]any{"profile": profile})
}
