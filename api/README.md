# Nexus API (Skeleton)

## Requisitos
- Go 1.23+

## Variables de entorno
Copiar `.env.example` y ajustar valores:
- `APP_ENV`
- `PORT`
- `SUPABASE_JWT_SECRET`
- `DATABASE_URL` (opcional; activa `vault.Store` en Postgres)
- `SUPABASE_URL`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_STORAGE_BUCKET`
- `GEMINI_API_KEY` (opcional; activa extractor Gemini)
- `GEMINI_MODEL` (opcional)
- `IDEMPOTENCY_CLEANUP_INTERVAL_SECONDS` (opcional; 0 desactiva)
- `IDEMPOTENCY_CLEANUP_BATCH_SIZE` (opcional)
- `CHILECOMPRA_BASE_URL` (opcional; activa cliente real Radar)
- `CHILECOMPRA_API_KEY` (opcional)
- `CHILECOMPRA_TENDERS_PATH` (opcional)
- `TENDER_SCORE_CACHE_TTL_SECONDS` (opcional; default 900)

## Ejecutar servidor
```bash
go run ./cmd/server
```

## Ejecutar con Docker (sin Go local)
Desde `api/`:
```bash
make docker-go-version
make docker-go-tidy
make docker-go-test
```

## CI/CD (NXS-010 fase 1)
- Workflow CI: `.github/workflows/ci.yml`
- Gate minimo: tests Go en Docker para `push` a `main` y `pull_request`.
- Convencion release/changelog: `docs/NXS-010_RELEASE_CONVENTION.md` y `docs/TECH_CHANGELOG.md`.
- Enforcement remoto (branch protection/checks): `docs/NXS-010_REMOTE_ENFORCEMENT.md` y `scripts/setup_github_enforcement.sh`.
- Nota: en repo privado, branch protection puede requerir plan GitHub con esa capacidad.

## Endpoints base
- `GET /health/live`
- `GET /health/ready`
- `GET /metrics` (metricas Prometheus texto plano)
- `GET /v1/protected` (requiere Bearer JWT)
- `POST /v1/vault/upload` (requiere Bearer JWT)
- `POST /v1/vault/process` (requiere Bearer JWT, solo estado uploaded)
- `POST /v1/vault/items/{id}/retry` (requiere Bearer JWT, solo estado failed)
- `GET /v1/vault/items` (requiere Bearer JWT)
- `GET /v1/vault/items/{id}` (requiere Bearer JWT)
- `GET /v1/vault/items/{id}/events` (requiere Bearer JWT)
- `GET /v1/tenders/sync` (requiere Bearer JWT)
- `GET /v1/tenders` (requiere Bearer JWT)
- `GET /v1/tenders/{id}/score` (requiere Bearer JWT)
- `POST /v1/tenders/score/warmup` (requiere Bearer JWT)
- `GET /v1/company/profile` (requiere Bearer JWT)
- `PUT /v1/company/profile` (requiere Bearer JWT)
- `GET /v1/ops/alerts` (requiere Bearer JWT)

`GET /v1/vault/items/{id}/events` soporta filtros opcionales:
- `limit` (default 50, max 200)
- `event_type`
- `from` (RFC3339)
- `to` (RFC3339)
- `before_cursor` (cursor compuesto de paginacion descendente)
- `include_total=true` (opcional; agrega `total_count`)

La respuesta incluye `returned_count`, `has_more` y `next_cursor`.

Para `POST /v1/vault/process` y `POST /v1/vault/items/{id}/retry` puedes enviar header `Idempotency-Key` para replay seguro ante reintentos de red.
Los eventos de fallo de procesamiento incluyen `error_stage`, `error_category`, `retryable` y `document_family` para facilitar troubleshooting.
Los eventos `vault_item_processed` incluyen `key_fields` (`amount`, `date`, `company_name`), `key_fields_found` y `missing_key_fields`.

Si `IDEMPOTENCY_CLEANUP_INTERVAL_SECONDS > 0`, se ejecuta limpieza periodica de llaves de idempotencia expiradas.

Para `GET /v1/tenders/sync` puedes usar:
- `limit` (default 50, max 200)
- `since` (RFC3339)

Para `GET /v1/tenders/{id}/score` puedes usar:
- `company_region`
- `company_keywords` (separadas por coma)

Para `POST /v1/tenders/score/warmup` puedes enviar:
- body JSON opcional: `limit`, `company_region`, `company_keywords`
- body JSON opcional: `tender_ids` (lista de `external_id` para warmup selectivo)
- o query params equivalentes: `limit`, `company_region`, `company_keywords`
- query param opcional: `tender_ids` separado por coma
- maximo `200` `tender_ids` por request (exceso se reporta como omitido)
- respuesta incluye `processed_count`, `cache_hits`, `cache_writes`, `targeted_ids`, `skipped_ids`

Si no envias `company_region` o `company_keywords`, el score intenta usar el perfil guardado en `company/profile`.
La respuesta de score incluye `cache_hit`.
Al actualizar `company/profile`, el cache de score de esa empresa se invalida inmediatamente.

## Ejemplo upload
```bash
curl -X POST http://localhost:8080/v1/vault/upload \
  -H "Authorization: Bearer <JWT_SUPABASE>" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name":"bases_licitacion.pdf",
    "mime_type":"application/pdf",
    "size_bytes":123456,
    "sha256":"abc123..."
  }'
```

## Ejemplo process
```bash
curl -X POST http://localhost:8080/v1/vault/process \
  -H "Authorization: Bearer <JWT_SUPABASE>" \
  -H "Content-Type: application/json" \
  -d '{
    "item_id":"<ITEM_ID>"
  }'
```

Nota: si configuras `SUPABASE_URL` y `SUPABASE_SERVICE_ROLE_KEY`, el endpoint de upload intentara generar signed URL real en Supabase Storage. Si no estan definidos, usa signer placeholder.

Si configuras `DATABASE_URL`, la API intenta usar Postgres como store de boveda. Si la conexion falla, cae automaticamente a in-memory store.

Si configuras `GEMINI_API_KEY`, el procesamiento usa Gemini para clasificacion y texto extraido. Para `image/*`, `audio/*` y `application/pdf` se adjunta contenido binario (inline_data) con limite de tamano. Si no esta configurado, usa extractor simulado.
Con `GEMINI_API_KEY` activo, el pipeline usa matriz de fallback por familia documental: `gemini -> simulated -> manual_review_required` para `pdf/image/audio`.

Con credenciales Supabase (`SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY`), `POST /v1/vault/process` intenta leer el objeto real desde Storage antes de extraer contenido.

## Observabilidad minima (NXS-009 fase 1)
- Metricas HTTP:
  - `nexus_http_requests_total`
  - `nexus_http_request_duration_ms_sum`
  - `nexus_http_request_duration_ms_count`
- Metricas pipeline boveda:
  - `nexus_vault_processing_total` (labels: `result`, `document_family`, `error_category`)
  - `nexus_vault_inflight`
- Metricas warmup score:
  - `nexus_tenders_warmup_runs_total`
  - `nexus_tenders_warmup_processed_total`
  - `nexus_tenders_warmup_cache_hits_total`
  - `nexus_tenders_warmup_cache_writes_total`
  - `nexus_tenders_warmup_skipped_total`
- Alertas operativas base (`GET /v1/ops/alerts`):
  - `http_error_rate_high` (warning)
  - `vault_timeout_rate_high` (warning)
  - `vault_inflight_high` (critical)
- Dashboard compartible:
  - `docs/ops/NXS-009_DASHBOARD_MINIMO.md`
  - `docs/ops/NXS-009_grafana_dashboard.json`
- Playbook de respuesta:
  - `docs/ops/NXS-009_ALERT_PLAYBOOK.md`

## Evaluacion base de calidad documental (NXS-008)
- Dataset mini: `api/testdata/vault_extraction_dataset.json`
- Test de accuracy + umbral inicial (85%): `go test ./internal/vault -run TestDocumentTypeDatasetAccuracy -v`
- Test de clasificacion de errores del pipeline: `go test ./internal/vault -run TestClassifyProcessingError -v`
- Dataset de campos clave: `api/testdata/vault_key_fields_dataset.json`
- Test de cobertura campos clave + umbral inicial (80%): `go test ./internal/vault -run TestKeyFieldsCoverageDataset -v`

## Validacion E2E preproduccion
- Script smoke: `api/scripts/e2e_preprod_smoke.sh`
- Checklist/criterios: `docs/E2E_PREPROD_VALIDATION.md`
