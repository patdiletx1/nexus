# Nexus - Technical Changelog

## v0.1.0 - 2026-04-14
### Changes
- Baseline backend Go multi-tenant with auth, vault processing, idempotency, tenders, scoring and cache.
- NXS-008 hardening completed (error taxonomy, quality datasets, fallback matrix).
- NXS-009 observability completed (metrics, alerts endpoint, dashboard assets, runbook).

### Validation
- `go test ./...` ejecutado en Docker.

### Residual Risks
- Validacion E2E pendiente con credenciales reales (Supabase/Gemini/ChileCompra).
- NXS-010 pendiente de consolidar politica de release/tag con remoto.
