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

## 2026-04-14 - NXS-009 fase 2 (alertas operativas)
- **Autor agente:** Codex (Cursor)
- **Contexto:** complementar observabilidad con alertas runtime consumibles por API.
- **Cambios principales:**
  - Se agrego evaluacion de alertas sobre error rate HTTP, timeout de pipeline y backlog inflight.
  - Se incorporo gauge `nexus_vault_inflight` y tracking de jobs en procesamiento.
  - Se habilito endpoint autenticado `GET /v1/ops/alerts`.
- **Archivos clave:**
  - `api/internal/observability/metrics.go`
  - `api/internal/http/handlers/ops.go`
  - `api/internal/http/handlers/ops_test.go`
  - `api/internal/http/router.go`
  - `api/internal/http/handlers/vault.go`
  - `api/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
- **Validacion:**
  - Tests de alertas y render de metricas + suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Falta dashboard compartible para cierre completo de NXS-009.
- **Siguiente paso recomendado:**
  - Definir template de dashboard y playbook de respuesta operativa.

## 2026-04-14 - NXS-009 fase 3 (dashboard + runbook)
- **Autor agente:** Codex (Cursor)
- **Contexto:** cerrar observabilidad SRE-lite con activos compartibles para operacion diaria.
- **Cambios principales:**
  - Se agrego guia de dashboard minimo con queries PromQL para seguimiento diario.
  - Se agrego JSON importable de dashboard Grafana inicial.
  - Se agrego playbook de respuesta para alertas `http_error_rate`, `vault_timeout_rate` y `vault_inflight`.
- **Archivos clave:**
  - `docs/ops/NXS-009_DASHBOARD_MINIMO.md`
  - `docs/ops/NXS-009_grafana_dashboard.json`
  - `docs/ops/NXS-009_ALERT_PLAYBOOK.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
  - `docs/AGENT_HANDOFF_INDEX.md`
  - `api/README.md`
- **Validacion:**
  - Revision de consistencia entre metricas expuestas, alertas activas y formulas de dashboard.
- **Riesgos/pendientes:**
  - Ajustar umbrales con telemetria real una vez disponible trafico productivo.
- **Siguiente paso recomendado:**
  - Iniciar NXS-010 con CI obligatorio y convencion de release/changelog tecnico.

## 2026-04-14 - NXS-010 fase 1 (CI + release convention)
- **Autor agente:** Codex (Cursor)
- **Contexto:** iniciar quality gate formal para merges y trazabilidad tecnica por iteracion.
- **Cambios principales:**
  - Se agrego workflow CI para ejecutar tests Go en Docker en `push`/`pull_request`.
  - Se agrego plantilla de Pull Request con checklist de validacion.
  - Se agrego convencion de release y changelog tecnico base.
- **Archivos clave:**
  - `.github/workflows/ci.yml`
  - `.github/pull_request_template.md`
  - `docs/NXS-010_RELEASE_CONVENTION.md`
  - `docs/TECH_CHANGELOG.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
  - `api/README.md`
- **Validacion:**
  - Revisión estática de flujo CI y consistencia documental de gates.
- **Riesgos/pendientes:**
  - Falta configurar branch protection/checks requeridos en remoto.
- **Siguiente paso recomendado:**
  - Aplicar required checks y politica de merge en el repositorio remoto.

## 2026-04-14 - NXS-010 fase 2 (remote enforcement automation)
- **Autor agente:** Codex (Cursor)
- **Contexto:** preparar cierre de NXS-010 ante bloqueo por falta de autenticacion `gh`.
- **Cambios principales:**
  - Se agrego script ejecutable para crear repo remoto y aplicar branch protection/checks.
  - Se agrego guia rapida de enforcement remoto con pasos de validacion.
  - Se actualizaron docs de estado para indicar paso operativo exacto pendiente.
- **Archivos clave:**
  - `scripts/setup_github_enforcement.sh`
  - `docs/NXS-010_REMOTE_ENFORCEMENT.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
  - `docs/AGENT_HANDOFF_INDEX.md`
  - `api/README.md`
- **Validacion:**
  - Verificacion de prerequisito `gh auth status`: actualmente sin sesion iniciada.
- **Riesgos/pendientes:**
  - Sin login de GitHub no se puede aplicar protection en remoto.
- **Siguiente paso recomendado:**
  - Ejecutar `gh auth login` y correr `./scripts/setup_github_enforcement.sh nexus private`.

## 2026-04-14 - NXS-010 fase 3 (repo remoto creado, protection bloqueada)
- **Autor agente:** Codex (Cursor)
- **Contexto:** ejecutar cierre remoto real tras login GitHub CLI.
- **Cambios principales:**
  - Se creo repositorio remoto `github.com/patdiletx1/nexus` y se publico `main`.
  - Intento de branch protection devolvio `403` por limitacion de plan en repo privado.
  - Se robustecio script para reportar esta condicion y sugerir opciones de salida.
- **Archivos clave:**
  - `scripts/setup_github_enforcement.sh`
  - `docs/NXS-010_REMOTE_ENFORCEMENT.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
- **Validacion:**
  - `gh auth status` valido.
  - ejecucion de script con creacion remota exitosa y evidencia de bloqueo en protection.
- **Riesgos/pendientes:**
  - NXS-010 queda bloqueado hasta definir visibilidad/plan para branch protection.
- **Siguiente paso recomendado:**
  - Decidir: upgrade plan o repo publico; luego rerun de `./scripts/setup_github_enforcement.sh`.

## 2026-04-14 - NXS-010 fase 4 (enforcement remoto aplicado)
- **Autor agente:** Codex (Cursor)
- **Contexto:** cerrar definitivamente NXS-010 tras resolver bloqueo de branch protection.
- **Cambios principales:**
  - Se actualizo visibilidad del repo a publico para habilitar protection por plan.
  - Se aplico branch protection en `main` con check requerido `backend-go-tests`.
  - Se actualizaron documentos de estado marcando NXS-010 en `done`.
- **Archivos clave:**
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
  - `docs/NXS-010_REMOTE_ENFORCEMENT.md`
  - `docs/AGENT_CHANGELOG.md`
- **Validacion:**
  - Script `./scripts/setup_github_enforcement.sh nexus public` ejecutado con resultado exitoso.
  - Verificacion repo `patdiletx1/nexus` en visibilidad publica.
- **Riesgos/pendientes:**
  - Pendiente principal desplazado a validacion E2E con credenciales reales de proveedores.
- **Siguiente paso recomendado:**
  - Ejecutar plan E2E controlado y documentar resultados contra criterios de pre-produccion.

## 2026-04-14 - Score cache warmup endpoint
- **Autor agente:** Codex (Cursor)
- **Contexto:** acelerar UX de listados de licitaciones precalentando score cache por empresa.
- **Cambios principales:**
  - Se implemento `POST /v1/tenders/score/warmup`.
  - Warmup usa inputs directos o fallback a `company/profile` para construir fingerprint.
  - Se agregaron tests de cache writes/hits y fallback de perfil.
- **Archivos clave:**
  - `api/internal/http/handlers/tenders.go`
  - `api/internal/http/handlers/tenders_test.go`
  - `api/internal/http/router.go`
  - `api/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
- **Validacion:**
  - Suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Pendiente medir impacto real de warmup en latencia de listados con trafico productivo.
- **Siguiente paso recomendado:**
  - Integrar llamada de warmup en flujo frontend al entrar al radar.
