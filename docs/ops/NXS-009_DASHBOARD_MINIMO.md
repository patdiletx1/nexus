# NXS-009 - Dashboard Operativo Minimo

Dashboard compartible para monitoreo diario del backend Nexus.

## Fuente de datos
- Endpoint Prometheus scrape target: `GET /metrics`
- Periodo sugerido: 5m para lectura operativa, 1h para tendencia.

## Paneles recomendados

### 1) HTTP Error Rate (%)
**Objetivo:** detectar degradacion de API.

**Formula sugerida**
```promql
100 * sum(rate(nexus_http_requests_total{status=~"5.."}[5m])) / sum(rate(nexus_http_requests_total[5m]))
```

### 2) HTTP P95 aproximado por endpoint (ms)
**Objetivo:** detectar latencias anormales.

**Nota:** por ahora tenemos `sum/count`; usar promedio por endpoint como aproximacion inicial.

**Formula sugerida**
```promql
sum by (path) (rate(nexus_http_request_duration_ms_sum[5m])) / sum by (path) (rate(nexus_http_request_duration_ms_count[5m]))
```

### 3) Vault Timeout Rate (%)
**Objetivo:** vigilar estabilidad de pipeline documental.

**Formula sugerida**
```promql
100 * sum(rate(nexus_vault_processing_total{error_category="timeout"}[5m])) / sum(rate(nexus_vault_processing_total[5m]))
```

### 4) Vault Processing Outcomes (stacked)
**Objetivo:** ver proporcion de `processed` vs `failed`.

**Formula sugerida**
```promql
sum by (result, document_family) (rate(nexus_vault_processing_total[5m]))
```

### 5) Vault Inflight Jobs
**Objetivo:** detectar backlog operativo.

**Formula sugerida**
```promql
nexus_vault_inflight
```

## Umbrales sugeridos (alineados a alertas)
- HTTP error rate: warning >= 5%
- Vault timeout rate: warning >= 20%
- Vault inflight: critical >= 10

## Operacion diaria
- Revisión al inicio del día: 5 minutos.
- Revisión post-deploy: 10 minutos.
- Si un umbral se dispara, seguir `docs/ops/NXS-009_ALERT_PLAYBOOK.md`.
