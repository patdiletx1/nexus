package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func Liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "live",
		Timestamp: time.Now().UTC(),
	})
}

func Readiness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC(),
	})
}

func ProtectedExample(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "authenticated request",
		"user": map[string]string{
			"user_id":    UserIDFromContext(r.Context()),
			"company_id": CompanyIDFromContext(r.Context()),
			"role":       RoleFromContext(r.Context()),
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
