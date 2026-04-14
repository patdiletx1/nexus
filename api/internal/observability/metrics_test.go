package observability

import (
	"strings"
	"testing"
)

func TestRenderPrometheusIncludesRecordedMetrics(t *testing.T) {
	m := NewMetrics()
	m.RecordHTTPRequest("GET", "/health/live", 200, 12)
	m.RecordVaultProcessing("failed", "pdf", "timeout")
	m.IncVaultInflight()

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
	if !strings.Contains(rendered, `nexus_vault_inflight 1`) {
		t.Fatalf("expected inflight metric, got:\n%s", rendered)
	}
}

func TestEvaluateAlerts(t *testing.T) {
	m := NewMetrics()

	for i := 0; i < 8; i++ {
		m.RecordHTTPRequest("GET", "/ok", 200, 5)
	}
	for i := 0; i < 2; i++ {
		m.RecordHTTPRequest("GET", "/err", 500, 5)
	}
	m.RecordVaultProcessing("failed", "pdf", "timeout")
	m.RecordVaultProcessing("processed", "pdf", "")
	m.IncVaultInflight()
	m.IncVaultInflight()
	m.IncVaultInflight()

	alerts := m.EvaluateAlerts(AlertThresholds{
		HTTPErrorRatePercent: 10,
		VaultTimeoutPercent:  30,
		VaultInflightMax:     2,
	})
	if len(alerts) != 3 {
		t.Fatalf("expected 3 alerts, got %d", len(alerts))
	}
	if !alerts[0].Triggered {
		t.Fatal("expected http error rate alert to trigger")
	}
	if !alerts[1].Triggered {
		t.Fatal("expected vault timeout alert to trigger")
	}
	if !alerts[2].Triggered {
		t.Fatal("expected inflight alert to trigger")
	}
}
