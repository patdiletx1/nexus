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
	vaultInflight   int64
	tenderWarmup    map[string]tenderWarmupCounter
}

type tenderWarmupCounter struct {
	Runs        int64
	Processed   int64
	CacheHits   int64
	CacheWrites int64
	Skipped     int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		httpRequests:       map[string]int64{},
		httpLatencyCount:   map[string]int64{},
		httpLatencyMsTotal: map[string]int64{},
		vaultProcessing:    map[string]int64{},
		vaultInflight:      0,
		tenderWarmup:       map[string]tenderWarmupCounter{},
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

func (m *Metrics) IncVaultInflight() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.vaultInflight++
}

func (m *Metrics) DecVaultInflight() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.vaultInflight > 0 {
		m.vaultInflight--
	}
}

func (m *Metrics) RecordTenderWarmup(profileSource, targetMode string, processed, cacheHits, cacheWrites, skipped int) {
	if m == nil {
		return
	}
	key := fmt.Sprintf(
		"profile_source=%s|target_mode=%s",
		sanitizeLabel(profileSource),
		sanitizeLabel(targetMode),
	)

	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.tenderWarmup[key]
	entry.Runs++
	entry.Processed += int64(processed)
	entry.CacheHits += int64(cacheHits)
	entry.CacheWrites += int64(cacheWrites)
	entry.Skipped += int64(skipped)
	m.tenderWarmup[key] = entry
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

	b.WriteString("# HELP nexus_vault_inflight Current in-flight vault processing jobs.\n")
	b.WriteString("# TYPE nexus_vault_inflight gauge\n")
	b.WriteString(fmt.Sprintf("nexus_vault_inflight %d\n", m.vaultInflight))

	b.WriteString("# HELP nexus_tenders_warmup_runs_total Total warmup runs by input mode.\n")
	b.WriteString("# TYPE nexus_tenders_warmup_runs_total counter\n")
	for _, key := range sortedWarmupKeys(m.tenderWarmup) {
		labels := parseLabels(key)
		entry := m.tenderWarmup[key]
		b.WriteString(fmt.Sprintf(
			"nexus_tenders_warmup_runs_total{profile_source=\"%s\",target_mode=\"%s\"} %d\n",
			labels["profile_source"],
			labels["target_mode"],
			entry.Runs,
		))
	}

	b.WriteString("# HELP nexus_tenders_warmup_processed_total Total tenders processed in warmup.\n")
	b.WriteString("# TYPE nexus_tenders_warmup_processed_total counter\n")
	for _, key := range sortedWarmupKeys(m.tenderWarmup) {
		labels := parseLabels(key)
		entry := m.tenderWarmup[key]
		b.WriteString(fmt.Sprintf(
			"nexus_tenders_warmup_processed_total{profile_source=\"%s\",target_mode=\"%s\"} %d\n",
			labels["profile_source"],
			labels["target_mode"],
			entry.Processed,
		))
	}

	b.WriteString("# HELP nexus_tenders_warmup_cache_hits_total Total warmup cache hits.\n")
	b.WriteString("# TYPE nexus_tenders_warmup_cache_hits_total counter\n")
	for _, key := range sortedWarmupKeys(m.tenderWarmup) {
		labels := parseLabels(key)
		entry := m.tenderWarmup[key]
		b.WriteString(fmt.Sprintf(
			"nexus_tenders_warmup_cache_hits_total{profile_source=\"%s\",target_mode=\"%s\"} %d\n",
			labels["profile_source"],
			labels["target_mode"],
			entry.CacheHits,
		))
	}

	b.WriteString("# HELP nexus_tenders_warmup_cache_writes_total Total warmup cache writes.\n")
	b.WriteString("# TYPE nexus_tenders_warmup_cache_writes_total counter\n")
	for _, key := range sortedWarmupKeys(m.tenderWarmup) {
		labels := parseLabels(key)
		entry := m.tenderWarmup[key]
		b.WriteString(fmt.Sprintf(
			"nexus_tenders_warmup_cache_writes_total{profile_source=\"%s\",target_mode=\"%s\"} %d\n",
			labels["profile_source"],
			labels["target_mode"],
			entry.CacheWrites,
		))
	}

	b.WriteString("# HELP nexus_tenders_warmup_skipped_total Total warmup skipped IDs.\n")
	b.WriteString("# TYPE nexus_tenders_warmup_skipped_total counter\n")
	for _, key := range sortedWarmupKeys(m.tenderWarmup) {
		labels := parseLabels(key)
		entry := m.tenderWarmup[key]
		b.WriteString(fmt.Sprintf(
			"nexus_tenders_warmup_skipped_total{profile_source=\"%s\",target_mode=\"%s\"} %d\n",
			labels["profile_source"],
			labels["target_mode"],
			entry.Skipped,
		))
	}

	return b.String()
}

type AlertThresholds struct {
	HTTPErrorRatePercent float64
	VaultTimeoutPercent  float64
	VaultInflightMax     int64
}

type Alert struct {
	Name        string  `json:"name"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Triggered   bool    `json:"triggered"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
}

func (m *Metrics) EvaluateAlerts(thresholds AlertThresholds) []Alert {
	if m == nil {
		return []Alert{}
	}

	if thresholds.HTTPErrorRatePercent <= 0 {
		thresholds.HTTPErrorRatePercent = 5
	}
	if thresholds.VaultTimeoutPercent <= 0 {
		thresholds.VaultTimeoutPercent = 20
	}
	if thresholds.VaultInflightMax <= 0 {
		thresholds.VaultInflightMax = 10
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	totalRequests := int64(0)
	errorRequests := int64(0)
	for key, count := range m.httpRequests {
		labels := parseLabels(key)
		totalRequests += count
		if strings.HasPrefix(labels["status"], "5") {
			errorRequests += count
		}
	}
	httpErrorRate := 0.0
	if totalRequests > 0 {
		httpErrorRate = float64(errorRequests) * 100 / float64(totalRequests)
	}

	totalVault := int64(0)
	timeoutVault := int64(0)
	for key, count := range m.vaultProcessing {
		labels := parseLabels(key)
		totalVault += count
		if labels["error_category"] == "timeout" {
			timeoutVault += count
		}
	}
	timeoutRate := 0.0
	if totalVault > 0 {
		timeoutRate = float64(timeoutVault) * 100 / float64(totalVault)
	}

	inflight := float64(m.vaultInflight)
	alerts := []Alert{
		{
			Name:        "http_error_rate_high",
			Severity:    "warning",
			Description: "HTTP 5xx rate is above threshold",
			Triggered:   httpErrorRate >= thresholds.HTTPErrorRatePercent,
			Value:       httpErrorRate,
			Threshold:   thresholds.HTTPErrorRatePercent,
		},
		{
			Name:        "vault_timeout_rate_high",
			Severity:    "warning",
			Description: "Vault processing timeout rate is above threshold",
			Triggered:   timeoutRate >= thresholds.VaultTimeoutPercent,
			Value:       timeoutRate,
			Threshold:   thresholds.VaultTimeoutPercent,
		},
		{
			Name:        "vault_inflight_high",
			Severity:    "critical",
			Description: "Vault in-flight processing jobs exceed safe threshold",
			Triggered:   inflight >= float64(thresholds.VaultInflightMax),
			Value:       inflight,
			Threshold:   float64(thresholds.VaultInflightMax),
		},
	}
	return alerts
}

func sortedKeys(values map[string]int64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedWarmupKeys(values map[string]tenderWarmupCounter) []string {
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
