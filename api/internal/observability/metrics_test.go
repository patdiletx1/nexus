package observability

import (
	"strings"
	"testing"
)

func TestRenderPrometheusIncludesRecordedMetrics(t *testing.T) {
	m := NewMetrics()
	m.RecordHTTPRequest("GET", "/health/live", 200, 12)
	m.RecordVaultProcessing("failed", "pdf", "timeout")

	rendered := m.RenderPrometheus()
	if !strings.Contains(rendered, `nexus_http_requests_total{method="GET",path="/health/live",status="200"} 1`) {
		t.Fatalf("expected http requests metric, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, `nexus_http_request_duration_ms_sum{method="GET",path="/health/live"} 12`) {
		t.Fatalf("expected duration sum metric, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, `nexus_vault_processing_total{result="failed",document_family="pdf",error_category="timeout"} 1`) {
		t.Fatalf("expected vault metric, got:\n%s", rendered)
	}
}
