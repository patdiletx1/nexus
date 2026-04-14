# Nexus RFP - Agent Changelog

Registro operativo de sesiones para continuidad entre agentes.

## Como usar este archivo
- Agregar una nueva entrada por sesion relevante.
- Mantener foco en hechos verificables, no en planes abstractos.
- Enlazar siempre:
  - cambios realizados,
  - validacion ejecutada,
  - pendientes y riesgos.

## Formato de entrada
```markdown
## YYYY-MM-DD - <titulo corto>
- **Autor agente:** <nombre/agente>
- **Contexto:** <que tarea se tomo>
- **Cambios principales:**
  - <cambio 1>
  - <cambio 2>
- **Archivos clave:**
  - `<ruta/archivo>`
  - `<ruta/archivo>`
- **Validacion:**
  - <tests/comandos y resultado>
- **Riesgos/pendientes:**
  - <riesgo o gap>
- **Siguiente paso recomendado:**
  - <accion concreta>
```

---

## 2026-04-13 - Base backend y esquema multi-tenant
- **Autor agente:** Codex (Cursor)
- **Contexto:** ejecucion inicial de Sprint 1 para habilitar base operacional.
- **Cambios principales:**
  - Se implemento API Go base con health checks, middleware JWT y logging con `request_id`.
  - Se crearon migraciones iniciales con tablas core y politicas RLS multi-tenant.
  - Se habilito estructura de handlers/rutas para crecimiento incremental.
- **Archivos clave:**
  - `api/cmd/server/main.go`
  - `api/internal/http/router.go`
  - `api/internal/http/middleware/auth.go`
  - `supabase/migrations/0001_init.sql`
- **Validacion:**
  - Pruebas unitarias de auth middleware y smoke de endpoints base.
- **Riesgos/pendientes:**
  - Falta validar comportamiento completo con credenciales reales de entorno productivo.
- **Siguiente paso recomendado:**
  - Completar flujo de boveda documental con estados y trazabilidad.

## 2026-04-13 - Boveda documental y pipeline asincrono inicial
- **Autor agente:** Codex (Cursor)
- **Contexto:** cubrir NXS-003/NXS-004 con flujo completo upload -> procesamiento.
- **Cambios principales:**
  - Se implementaron endpoints de boveda: upload, process, retry, list, detalle y eventos.
  - Se agrego procesamiento asincrono con extractor simulado y extractor Gemini opcional.
  - Se incorporo lectura de objetos desde storage y persistencia de texto extraido/tipo documental/error.
- **Archivos clave:**
  - `api/internal/http/handlers/vault.go`
  - `api/internal/vault/store.go`
  - `api/internal/vault/postgres_store.go`
  - `api/internal/vault/gemini_extractor.go`
  - `supabase/migrations/0002_vault_processing_fields.sql`
- **Validacion:**
  - Suite de tests de handlers de boveda y tests de helpers del extractor Gemini.
- **Riesgos/pendientes:**
  - Hardening de calidad OCR/extraction pendiente (dataset y umbrales formales).
- **Siguiente paso recomendado:**
  - Fortalecer manejo de errores por tipo documental y precision por campo.

## 2026-04-13 - Auditoria, idempotencia y robustez operacional
- **Autor agente:** Codex (Cursor)
- **Contexto:** reducir reprocesos y mejorar trazabilidad de estados en operaciones async.
- **Cambios principales:**
  - Se implemento servicio de idempotencia con almacenamiento en Postgres y cleanup por expiracion.
  - Se incorporaron filtros y paginacion robusta en auditoria de eventos por item.
  - Se ajusto orden de validacion en `retry` para respetar replay idempotente.
- **Archivos clave:**
  - `api/internal/idempotency/postgres.go`
  - `api/internal/http/handlers/vault.go`
  - `api/internal/audit/postgres_logger.go`
  - `supabase/migrations/0003_audit_events_query_indexes.sql`
  - `supabase/migrations/0004_idempotency_keys.sql`
- **Validacion:**
  - Tests de idempotencia replay y conflictos de estado en `vault_test`.
- **Riesgos/pendientes:**
  - Falta observabilidad de metricas/alertas (solo logs estructurados en esta etapa).
- **Siguiente paso recomendado:**
  - Implementar capa de metricas por endpoint y alertas operativas base.

## 2026-04-14 - Radar ChileCompra, scoring y cache por empresa
- **Autor agente:** Codex (Cursor)
- **Contexto:** avanzar Sprint 2 en captura de oportunidades y priorizacion comercial.
- **Cambios principales:**
  - Se implemento cliente ChileCompra configurable con sync/list de tenders.
  - Se agrego endpoint de score explicable por licitacion con perfil de empresa persistente.
  - Se incorporo cache de score (in-memory/Postgres) e invalidacion al actualizar perfil.
- **Archivos clave:**
  - `api/internal/chilecompra/client.go`
  - `api/internal/http/handlers/tenders.go`
  - `api/internal/http/handlers/company_profile.go`
  - `api/internal/tenders/postgres_score_cache.go`
  - `supabase/migrations/0005_tenders.sql`
  - `supabase/migrations/0006_company_scoring_profiles.sql`
  - `supabase/migrations/0007_tender_score_cache.sql`
- **Validacion:**
  - Tests de cliente ChileCompra, handlers de tenders y profile invalidando cache.
  - `go test ./...` ejecutado en entorno Docker.
- **Riesgos/pendientes:**
  - Validacion final con credenciales/productor real ChileCompra pendiente.
- **Siguiente paso recomendado:**
  - Implementar endpoint `POST /v1/tenders/score/warmup` para precalentar cache por empresa.

## 2026-04-14 - Setup de handoff y estado
- **Autor agente:** Codex (Cursor)
- **Contexto:** consolidar documentacion de continuidad entre agentes.
- **Cambios principales:**
  - Se creo set de handoff en `docs/` con indice, estado y playbook.
  - Se agrego backlog operativo actualizado y recomendaciones de siguiente tarea.
- **Archivos clave:**
  - `docs/AGENT_HANDOFF_INDEX.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/AGENT_EXECUTION_PLAYBOOK.md`
  - `docs/nexus-sprint-01-backlog.md`
- **Validacion:**
  - Revision de consistencia documental y estructura de lectura recomendada.
- **Riesgos/pendientes:**
  - Riesgo de desalineacion si no se actualiza changelog al cierre de cada bloque.
- **Siguiente paso recomendado:**
  - Registrar cada nueva iteracion tecnica usando este formato y actualizar estado global.

## 2026-04-14 - Alineacion de estado Sprint 2
- **Autor agente:** Codex (Cursor)
- **Contexto:** corregir discrepancias entre backlog y estado de handoff.
- **Cambios principales:**
  - Se alinearon estados de Sprint 2: NXS-006/NXS-007 en `done`.
  - Se ajusto NXS-008 a `next` y se reordeno plan recomendado a NXS-008 -> NXS-009 -> NXS-010.
  - Se agrego bloque explicito "Estado Sprint 2 (historias)" en `AGENT_PROJECT_STATUS`.
- **Archivos clave:**
  - `docs/nexus-sprint-01-backlog.md`
  - `docs/AGENT_PROJECT_STATUS.md`
- **Validacion:**
  - Revision cruzada manual entre estado global, backlog y tarea recomendada.
- **Riesgos/pendientes:**
  - Mantener consistencia futura entre docs al cerrar cada historia.
- **Siguiente paso recomendado:**
  - Iniciar NXS-008 con dataset mini y reporte de precision base.

## 2026-04-14 - NXS-008 fase 1 (errores + dataset base)
- **Autor agente:** Codex (Cursor)
- **Contexto:** iniciar hardening documental con foco en trazabilidad de fallos y medicion minima de calidad.
- **Cambios principales:**
  - Se implemento clasificacion de errores de procesamiento por `stage`, `category` y `retryable`.
  - Se enriquecio auditoria/log en fallos y exitos con `document_family` para analisis por tipo de entrada.
  - Se agrego dataset mini de evaluacion y test con umbral inicial de accuracy (85%) para `doc_type`.
- **Archivos clave:**
  - `api/internal/vault/error_classification.go`
  - `api/internal/vault/quality_dataset_test.go`
  - `api/internal/http/handlers/vault.go`
  - `api/internal/http/handlers/vault_test.go`
  - `api/testdata/vault_extraction_dataset.json`
  - `api/README.md`
- **Validacion:**
  - Tests unitarios de clasificacion de errores y accuracy de dataset documental.
  - Ajuste de test de handler para verificar payload de fallo enriquecido.
- **Riesgos/pendientes:**
  - Accuracy actual solo cubre `doc_type`; faltan metricas por campos de negocio (monto, fecha, razon social).
- **Siguiente paso recomendado:**
  - Extender evaluacion a campos clave y definir fallback por proveedor/tipo documental.

## 2026-04-14 - NXS-008 fase 2 (campos clave)
- **Autor agente:** Codex (Cursor)
- **Contexto:** ampliar medicion de calidad documental hacia campos de negocio utiles para propuesta.
- **Cambios principales:**
  - Se implemento extraccion de campos clave desde texto (`amount`, `date`, `company_name`).
  - Se agrego dataset de validacion de campos clave con umbral inicial de cobertura (80%).
  - Se enriquecio evento `vault_item_processed` con `key_fields`, `key_fields_found` y `missing_key_fields`.
- **Archivos clave:**
  - `api/internal/vault/key_fields.go`
  - `api/internal/vault/key_fields_test.go`
  - `api/internal/vault/key_fields_quality_test.go`
  - `api/internal/http/handlers/vault.go`
  - `api/internal/http/handlers/vault_test.go`
  - `api/testdata/vault_key_fields_dataset.json`
  - `api/README.md`
- **Validacion:**
  - Tests unitarios de extraccion de campos y cobertura por dataset.
  - Test de handler validando payload de evento procesado con campos clave.
- **Riesgos/pendientes:**
  - Falta matriz de fallback por proveedor/tipo documental para cierre de NXS-008.
- **Siguiente paso recomendado:**
  - Definir estrategia de fallback primario/secundario/manual por tipo documental.

## 2026-04-14 - NXS-008 fase 3 (fallback matrix runtime)
- **Autor agente:** Codex (Cursor)
- **Contexto:** cerrar hardening documental con degradacion controlada por proveedor y tipo de documento.
- **Cambios principales:**
  - Se implemento extractor con matriz de fallback por familia documental.
  - Para `pdf/image/audio` se aplica cadena `gemini -> simulated -> manual_review_required`.
  - Se cableo la matriz en `main` cuando existe `GEMINI_API_KEY`.
- **Archivos clave:**
  - `api/internal/vault/fallback_extractor.go`
  - `api/internal/vault/fallback_extractor_test.go`
  - `api/internal/vault/error_classification.go`
  - `api/cmd/server/main.go`
  - `api/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
- **Validacion:**
  - Tests unitarios del fallback extractor (primario, secundario y manual review).
  - Suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Ajustar estrategia final con data real de precision/costos por proveedor en ambiente productivo.
- **Siguiente paso recomendado:**
  - Iniciar NXS-009 con metricas y alertas sobre errores y timeouts del pipeline.

## 2026-04-14 - NXS-009 fase 1 (metricas base)
- **Autor agente:** Codex (Cursor)
- **Contexto:** iniciar observabilidad operativa con senales minimas de latencia y fallos.
- **Cambios principales:**
  - Se implemento colector in-memory de metricas con salida Prometheus (`GET /metrics`).
  - Se agrego instrumentacion HTTP (requests totales + suma/conteo de latencia).
  - Se agregaron metricas de pipeline de boveda por resultado/familia/error.
- **Archivos clave:**
  - `api/internal/observability/metrics.go`
  - `api/internal/observability/metrics_test.go`
  - `api/internal/http/router.go`
  - `api/internal/http/handlers/vault.go`
  - `api/cmd/server/main.go`
  - `api/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
- **Validacion:**
  - Tests unitarios de render Prometheus y suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Faltan alertas automáticas y dashboard para cierre completo de NXS-009.
- **Siguiente paso recomendado:**
  - Definir umbrales de alerta (error_rate/timeout/backlog) y playbook de respuesta.
