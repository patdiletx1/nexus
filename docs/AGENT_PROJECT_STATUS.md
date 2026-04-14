# Nexus RFP - Project Status (Handoff)

## Estado actual
- **Fase:** Sprint 2 (backend/radar en construccion).
- **Backend Go:** funcional y probado en Docker (`go test ./...`).
- **DB/RLS:** base multi-tenant implementada con migraciones.

## Estado Sprint 2 (historias)
- **NXS-006 (Radar ChileCompra):** `done` (pendiente validacion final con credenciales reales).
- **NXS-007 (Score inicial):** `done`.
- **NXS-008 (Hardening documental):** `done`.
- **NXS-009 (Observabilidad SRE-lite):** `in_progress` (metricas base HTTP y pipeline boveda implementadas).
- **NXS-010 (CI/CD y quality gate):** `next`.

## Hecho (resumen tecnico)
- Base API: auth JWT, health checks, logging con request id.
- Boveda: upload, process, retry, list, get, eventos.
- Pipeline documental: async, extractor simulado/Gemini, lectura desde Supabase Storage (si hay credenciales).
- Auditoria: eventos por item con filtros, cursor compuesto y paginacion robusta.
- Idempotencia: `Idempotency-Key` en process/retry + cleanup de expiradas.
- Radar base: sync/list de tenders con cliente ChileCompra configurable.
- Scoring inicial: endpoint de score explicable + perfil persistente por empresa + cache de score.
- Invalidacion de cache: al actualizar `company/profile`.
- Hardening documental (parcial): clasificacion de errores (`error_stage`, `error_category`, `retryable`) y dataset mini con umbral inicial para `doc_type`.
- Quality signals en procesamiento: extraccion de campos clave (`amount`, `date`, `company_name`) con cobertura base por dataset.
- Fallback runtime documental: matriz por familia (`pdf/image/audio`) con flujo `gemini -> simulated -> manual_review_required`.

## Pendiente principal
- Integracion real validada contra API ChileCompra en entorno con credenciales reales.
- Hardening OCR productivo con medicion de precision por campos clave (no solo doc_type) y fallback por proveedor.
- Observabilidad SRE-lite (metricas + alertas, no solo logs).
- CI/CD con gates formales para merge/release.
- Frontend Flutter (aun no abordado en este repo).

## Endpoints relevantes ya disponibles
- Boveda: `POST /v1/vault/upload`, `POST /v1/vault/process`, `POST /v1/vault/items/{id}/retry`
- Boveda lectura: `GET /v1/vault/items`, `GET /v1/vault/items/{id}`, `GET /v1/vault/items/{id}/events`
- Radar: `GET /v1/tenders/sync`, `GET /v1/tenders`, `GET /v1/tenders/{id}/score`
- Perfil de score empresa: `GET /v1/company/profile`, `PUT /v1/company/profile`

## Migraciones existentes
- `0001_init.sql`
- `0002_vault_processing_fields.sql`
- `0003_audit_events_query_indexes.sql`
- `0004_idempotency_keys.sql`
- `0005_tenders.sql`
- `0006_company_scoring_profiles.sql`
- `0007_tender_score_cache.sql`

## Siguiente tarea recomendada
Completar NXS-009 con alertas operativas (error_rate, timeout procesamiento, backlog worker) y playbook de respuesta.
