package observability

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Metrics struct {
	mu sync.RWMutex

	httpRequests       map[string]int64
	httpLatencyCount   map[string]int64
	httpLatencyMsTotal map[string]int64

	vaultProcessing map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		httpRequests:       map[string]int64{},
		httpLatencyCount:   map[string]int64{},
		httpLatencyMsTotal: map[string]int64{},
		vaultProcessing:    map[string]int64{},
	}
}

func (m *Metrics) RecordHTTPRequest(method, path string, statusCode int, durationMs int64) {
	if m == nil {
		return
	}

	requestKey := fmt.Sprintf("method=%s|path=%s|status=%d", sanitizeLabel(method), sanitizeLabel(path), statusCode)
	latencyKey := fmt.Sprintf("method=%s|path=%s", sanitizeLabel(method), sanitizeLabel(path))

	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpRequests[requestKey]++
	m.httpLatencyCount[latencyKey]++
	m.httpLatencyMsTotal[latencyKey] += durationMs
}

func (m *Metrics) RecordVaultProcessing(result, documentFamily, errorCategory string) {
	if m == nil {
		return
	}
	key := fmt.Sprintf(
		"result=%s|document_family=%s|error_category=%s",
		sanitizeLabel(result),
		sanitizeLabel(documentFamily),
		sanitizeLabel(errorCategory),
	)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.vaultProcessing[key]++
}

func (m *Metrics) RenderPrometheus() string {
	if m == nil {
		return ""
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var b strings.Builder
	b.WriteString("# HELP nexus_http_requests_total Total HTTP requests by status.\n")
	b.WriteString("# TYPE nexus_http_requests_total counter\n")
	for _, key := range sortedKeys(m.httpRequests) {
		labels := parseLabels(key)
		b.WriteString(fmt.Sprintf(
			"nexus_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
			labels["method"],
			labels["path"],
			labels["status"],
			m.httpRequests[key],
		))
	}

	b.WriteString("# HELP nexus_http_request_duration_ms_sum Sum of HTTP request durations in milliseconds.\n")
	b.WriteString("# TYPE nexus_http_request_duration_ms_sum counter\n")
	for _, key := range sortedKeys(m.httpLatencyMsTotal) {
		labels := parseLabels(key)
		b.WriteString(fmt.Sprintf(
			"nexus_http_request_duration_ms_sum{method=\"%s\",path=\"%s\"} %d\n",
			labels["method"],
			labels["path"],
			m.httpLatencyMsTotal[key],
		))
	}

	b.WriteString("# HELP nexus_http_request_duration_ms_count Count of HTTP request durations.\n")
	b.WriteString("# TYPE nexus_http_request_duration_ms_count counter\n")
	for _, key := range sortedKeys(m.httpLatencyCount) {
		labels := parseLabels(key)
		b.WriteString(fmt.Sprintf(
			"nexus_http_request_duration_ms_count{method=\"%s\",path=\"%s\"} %d\n",
			labels["method"],
			labels["path"],
			m.httpLatencyCount[key],
		))
	}

	b.WriteString("# HELP nexus_vault_processing_total Total vault processing outcomes.\n")
	b.WriteString("# TYPE nexus_vault_processing_total counter\n")
	for _, key := range sortedKeys(m.vaultProcessing) {
		labels := parseLabels(key)
		b.WriteString(fmt.Sprintf(
			"nexus_vault_processing_total{result=\"%s\",document_family=\"%s\",error_category=\"%s\"} %d\n",
			labels["result"],
			labels["document_family"],
			labels["error_category"],
			m.vaultProcessing[key],
		))
	}

	return b.String()
}

func sortedKeys(values map[string]int64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func parseLabels(labelString string) map[string]string {
	out := map[string]string{}
	parts := strings.Split(labelString, "|")
	for _, part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			continue
		}
		out[keyValue[0]] = keyValue[1]
	}
	return out
}

func sanitizeLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "none"
	}
	return strings.ReplaceAll(value, "\"", "'")
}
