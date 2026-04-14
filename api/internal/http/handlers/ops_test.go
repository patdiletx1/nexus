package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nexus/api/internal/observability"
)

func TestOpsAlerts(t *testing.T) {
	metrics := observability.NewMetrics()
	for i := 0; i < 10; i++ {
		metrics.RecordHTTPRequest("GET", "/v1/test", 500, 10)
	}

	handler := OpsHandler{Metrics: metrics}
	req := httptest.NewRequest(http.MethodGet, "/v1/ops/alerts", nil)
	rec := httptest.NewRecorder()
	handler.Alerts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	alerts, ok := payload["alerts"].([]any)
	if !ok {
		t.Fatalf("expected alerts array, got %T", payload["alerts"])
	}
	if len(alerts) == 0 {
		t.Fatal("expected at least one alert in response")
	}
}
