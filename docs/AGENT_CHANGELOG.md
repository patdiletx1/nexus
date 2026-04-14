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

## 2026-04-14 - Warmup selectivo por tender IDs
- **Autor agente:** Codex (Cursor)
- **Contexto:** permitir precalentado de cache dirigido para lotes pequenos del frontend.
- **Cambios principales:**
  - Warmup ahora acepta `tender_ids` para procesar solo licitaciones especificas.
  - Se agrega deduplicacion de IDs y omision de IDs inexistentes.
  - Se agregan tests para validar comportamiento de targeting.
- **Archivos clave:**
  - `api/internal/http/handlers/tenders.go`
  - `api/internal/http/handlers/tenders_test.go`
  - `api/README.md`
- **Validacion:**
  - Suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Medir estrategia optima de lotes (N IDs por llamada) segun trafico real.
- **Siguiente paso recomendado:**
  - Definir tamano de lote recomendado y policy de reintento en frontend.

## 2026-04-14 - Warmup hardening (max IDs + skipped_ids)
- **Autor agente:** Codex (Cursor)
- **Contexto:** robustecer warmup selectivo para uso seguro en frontend.
- **Cambios principales:**
  - Se agrega limite maximo de `200` IDs objetivo por request.
  - Se reportan IDs omitidos en `skipped_ids` (duplicados, exceso y no encontrados).
  - Se agregan tests para limite y reporte de omitidos.
- **Archivos clave:**
  - `api/internal/http/handlers/tenders.go`
  - `api/internal/http/handlers/tenders_test.go`
  - `api/README.md`
- **Validacion:**
  - Suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Ajustar limite maximo segun observabilidad real de latencia/carga.
- **Siguiente paso recomendado:**
  - Exponer recomendacion de lote en frontend y agregar retry/backoff por batch.

## 2026-04-14 - E2E preprod validation pack
- **Autor agente:** Codex (Cursor)
- **Contexto:** habilitar ejecucion guiada de validacion real con credenciales externas.
- **Cambios principales:**
  - Se agrego script smoke para validar health/profile/sync/warmup/score contra entorno real.
  - Se agrego checklist de preproduccion con criterios de salida y evidencia requerida.
  - Se enlazo el paquete E2E en handoff/status/README para continuidad operativa.
- **Archivos clave:**
  - `api/scripts/e2e_preprod_smoke.sh`
  - `docs/E2E_PREPROD_VALIDATION.md`
  - `docs/AGENT_HANDOFF_INDEX.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `api/README.md`
- **Validacion:**
  - Validacion sintactica del script (`bash -n`).
- **Riesgos/pendientes:**
  - La corrida depende de credenciales y datos reales del entorno.
- **Siguiente paso recomendado:**
  - Ejecutar smoke en staging/preprod y adjuntar evidencia en changelog tecnico.

## 2026-04-14 - Warmup observability metrics
- **Autor agente:** Codex (Cursor)
- **Contexto:** medir efectividad de score warmup para ajuste de lotes y costo.
- **Cambios principales:**
  - Se agrego instrumentacion de warmup en `/metrics` (runs, processed, cache hits, writes, skipped).
  - Warmup registra metrica por `profile_source` y `target_mode`.
  - Se agregaron tests en `observability` y `tenders` para validar exposicion/registro.
- **Archivos clave:**
  - `api/internal/observability/metrics.go`
  - `api/internal/observability/metrics_test.go`
  - `api/internal/http/handlers/tenders.go`
  - `api/internal/http/handlers/tenders_test.go`
  - `api/internal/http/router.go`
  - `api/README.md`
- **Validacion:**
  - Suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Definir umbrales de alerta sobre tasas de `skipped` en warmup.
- **Siguiente paso recomendado:**
  - Agregar alerta cuando skipped ratio de warmup supere umbral configurable.

## 2026-04-14 - Warmup skipped-ratio alert
- **Autor agente:** Codex (Cursor)
- **Contexto:** cerrar loop operativo entre metricas warmup y alertas accionables.
- **Cambios principales:**
  - Se agrego alerta `tenders_warmup_skipped_ratio_high` en evaluacion de `/v1/ops/alerts`.
  - Se incluyo umbral configurable `WarmupSkippedRatioPercent` (default 30%).
  - Se actualizaron playbook/dashboard con checks y formula de ratio.
- **Archivos clave:**
  - `api/internal/observability/metrics.go`
  - `api/internal/observability/metrics_test.go`
  - `docs/ops/NXS-009_ALERT_PLAYBOOK.md`
  - `docs/ops/NXS-009_DASHBOARD_MINIMO.md`
  - `docs/AGENT_PROJECT_STATUS.md`
- **Validacion:**
  - Suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Ajustar umbral en base a comportamiento real de clientes.
- **Siguiente paso recomendado:**
  - Exponer threshold via config para tunning por entorno.

## 2026-04-14 - Alert thresholds configurable by environment
- **Autor agente:** Codex (Cursor)
- **Contexto:** permitir tuning operativo por entorno sin cambios de codigo.
- **Cambios principales:**
  - Se agregaron variables de entorno para umbrales de alertas (`http`, `vault`, `warmup`).
  - `OpsHandler` ahora evalua alertas usando thresholds inyectados desde config.
  - Se actualizaron `README` y `.env.example` con defaults operativos.
- **Archivos clave:**
  - `api/internal/config/config.go`
  - `api/cmd/server/main.go`
  - `api/internal/http/router.go`
  - `api/internal/http/handlers/ops.go`
  - `api/.env.example`
  - `api/README.md`
- **Validacion:**
  - Suite `go test ./...` en Docker.
- **Riesgos/pendientes:**
  - Definir valores objetivo por ambiente (`dev/staging/prod`) con datos reales.
- **Siguiente paso recomendado:**
  - Documentar baseline de thresholds por entorno en runbook SRE.

## 2026-04-14 - E2E smoke evidence automation
- **Autor agente:** Codex (Cursor)
- **Contexto:** facilitar validacion preprod con evidencia estructurada reutilizable en handoff.
- **Cambios principales:**
  - `api/scripts/e2e_preprod_smoke.sh` ahora genera evidencia JSON timestamped por corrida.
  - Se agregan variables `EVIDENCE_DIR` y `EVIDENCE_BASENAME` para controlar destino/nombre de evidencia.
  - Se actualizaron guias (`README` + checklist E2E + estado proyecto) para estandarizar el guardado de evidencia.
- **Archivos clave:**
  - `api/scripts/e2e_preprod_smoke.sh`
  - `docs/E2E_PREPROD_VALIDATION.md`
  - `api/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
- **Validacion:**
  - `bash -n api/scripts/e2e_preprod_smoke.sh`
- **Riesgos/pendientes:**
  - La corrida real sigue dependiendo de `JWT_TOKEN` y credenciales productivas de integraciones externas.
- **Siguiente paso recomendado:**
  - Ejecutar smoke real, versionar evidencia JSON y adjuntar snapshot de `/metrics` y `/v1/ops/alerts`.

## 2026-04-14 - E2E observability snapshots in artifact
- **Autor agente:** Codex (Cursor)
- **Contexto:** cerrar evidencia preprod en un solo archivo para reducir pasos manuales y errores operativos.
- **Cambios principales:**
  - El smoke E2E ahora captura automaticamente `GET /metrics` y `GET /v1/ops/alerts`.
  - Ambos snapshots quedan embebidos en el JSON de evidencia (`responses.metrics` y `responses.ops_alerts`).
  - Se ajustaron guias de validacion/estado para usar evidencia unificada.
- **Archivos clave:**
  - `api/scripts/e2e_preprod_smoke.sh`
  - `docs/E2E_PREPROD_VALIDATION.md`
  - `api/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
- **Validacion:**
  - `bash -n api/scripts/e2e_preprod_smoke.sh`
- **Riesgos/pendientes:**
  - La corrida depende de endpoints operativos y token con permisos sobre `company_id` real.
- **Siguiente paso recomendado:**
  - Ejecutar smoke real y validar que `responses.ops_alerts.alerts` no presente `critical` persistente.

## 2026-04-14 - Local full smoke con Radar mock + fix parser
- **Autor agente:** Codex (Cursor)
- **Contexto:** habilitar validacion E2E completa en local sin bloqueo por credenciales ChileCompra.
- **Cambios principales:**
  - Se agrego `CHILECOMPRA_MOCK_ENABLED` para usar licitaciones mock en `sync` cuando no hay cliente real configurado.
  - Se corrigio bug en `api/scripts/e2e_preprod_smoke.sh` al parsear `list_payload` (fallaba por lectura de stdin en Python).
  - Se documento modo local en README/checklist y se valido smoke completo con evidencia generada.
- **Archivos clave:**
  - `api/internal/chilecompra/client.go`
  - `api/internal/config/config.go`
  - `api/cmd/server/main.go`
  - `api/scripts/e2e_preprod_smoke.sh`
  - `api/.env.example`
  - `docs/E2E_PREPROD_VALIDATION.md`
- **Validacion:**
  - `dockerized go test ./...` (ok)
  - `./api/scripts/e2e_preprod_smoke.sh` contra backend local con `CHILECOMPRA_MOCK_ENABLED=true` (ok)
- **Riesgos/pendientes:**
  - El smoke real con integracion ChileCompra aun depende de credenciales y comportamiento externo.
- **Siguiente paso recomendado:**
  - Repetir smoke en preprod real y comparar diferencias de score/warmup frente al modo mock local.

## 2026-04-14 - Frontend Flutter bootstrap inicial
- **Autor agente:** Codex (Cursor)
- **Contexto:** avanzar en el frente cliente sin bloquearse por credenciales preprod.
- **Cambios principales:**
  - Se creo carpeta `frontend/` con `pubspec.yaml`, `analysis_options.yaml` y `lib/main.dart`.
  - Se implemento pantalla bootstrap para probar `health`, `company/profile` y `tenders` contra backend local.
  - Se agrego `frontend/README.md` con pasos de ejecucion y variables esperadas.
- **Archivos clave:**
  - `frontend/pubspec.yaml`
  - `frontend/lib/main.dart`
  - `frontend/README.md`
  - `.gitignore`
- **Validacion:**
  - Verificacion de estructura y dependencias base del proyecto Flutter.
- **Riesgos/pendientes:**
  - SDK Flutter local bloqueado por ownership en `/opt/homebrew/share/flutter` (error `dubious ownership`).
- **Siguiente paso recomendado:**
  - Resolver setup local del SDK Flutter y ejecutar `flutter pub get && flutter run` en `frontend/`.

## 2026-04-14 - Script helper para JWT local
- **Autor agente:** Codex (Cursor)
- **Contexto:** simplificar pruebas frontend/backend en local sin regenerar JWT manualmente cada vez.
- **Cambios principales:**
  - Se agrego `scripts/gen_local_jwt.sh` para emitir JWT HS256 compatible con `SUPABASE_JWT_SECRET`.
  - Se documento uso rapido y override de claims en `frontend/README.md`.
- **Archivos clave:**
  - `scripts/gen_local_jwt.sh`
  - `frontend/README.md`
- **Validacion:**
  - Script ejecutable y generacion de token por consola.
- **Riesgos/pendientes:**
  - Token local es solo para entorno de desarrollo con `local-dev-secret`.
- **Siguiente paso recomendado:**
  - Consumir helper en flujo frontend para autofill de token en modo local.

## 2026-04-14 - Frontend refactor por capas + acciones Radar
- **Autor agente:** Codex (Cursor)
- **Contexto:** continuar el bootstrap frontend para facilitar pruebas funcionales sin crecer un `main.dart` monolitico.
- **Cambios principales:**
  - Refactor de frontend en capas: `pages/home_page.dart`, `services/nexus_api_client.dart`, `widgets/response_card.dart`.
  - La UI ahora cubre acciones `health`, `profile`, `sync`, `list`, `warmup` y `score` con `tender_id` editable.
  - Se actualizo `frontend/README.md` con endpoints, uso y estructura de carpetas.
- **Archivos clave:**
  - `frontend/lib/main.dart`
  - `frontend/lib/pages/home_page.dart`
  - `frontend/lib/services/nexus_api_client.dart`
  - `frontend/lib/widgets/response_card.dart`
  - `frontend/README.md`
- **Validacion:**
  - `flutter analyze` (ok).
- **Riesgos/pendientes:**
  - Falta capa de estado mas robusta (provider/bloc) y navegacion multi-pantalla para escalar UI.
- **Siguiente paso recomendado:**
  - Separar secciones en pantallas dedicadas (`Radar`, `Company Profile`, `Ops`) y agregar manejo de errores por dominio.

## 2026-04-14 - Persistencia local de configuracion frontend
- **Autor agente:** Codex (Cursor)
- **Contexto:** evitar repetir pegado manual de URL/token en cada reinicio de app durante pruebas locales.
- **Cambios principales:**
  - Se agrego `shared_preferences` al frontend.
  - `HomePage` ahora carga/guarda automaticamente `API_BASE_URL`, `JWT_TOKEN` y `Tender ID`.
  - Se muestra indicador visual corto mientras se restauran valores al iniciar.
- **Archivos clave:**
  - `frontend/pubspec.yaml`
  - `frontend/lib/pages/home_page.dart`
  - `frontend/README.md`
- **Validacion:**
  - `flutter pub get`
  - `flutter analyze`
- **Riesgos/pendientes:**
  - JWT local queda persistido en storage del navegador/dispositivo (aceptable solo para entorno dev).
- **Siguiente paso recomendado:**
  - Agregar boton "limpiar sesion local" para reset de token/URL desde UI.

## 2026-04-14 - Controles UX de sesion local frontend
- **Autor agente:** Codex (Cursor)
- **Contexto:** completar ciclo de usabilidad local tras agregar persistencia de configuracion.
- **Cambios principales:**
  - Se agrego toggle de visibilidad para `JWT_TOKEN` en el input principal.
  - Se agrego boton `Limpiar sesion local` para resetear storage local y respuestas en UI.
  - Se actualizaron docs de frontend y estado de proyecto con estos controles.
- **Archivos clave:**
  - `frontend/lib/pages/home_page.dart`
  - `frontend/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
- **Validacion:**
  - `flutter analyze`
- **Riesgos/pendientes:**
  - Falta confirmacion explicita antes de limpiar sesion para prevenir accion accidental.
- **Siguiente paso recomendado:**
  - Agregar dialogo de confirmacion y snackbar de resultado al limpiar sesion.

## 2026-04-14 - Confirmacion y feedback en limpieza de sesion
- **Autor agente:** Codex (Cursor)
- **Contexto:** reducir riesgo de borrado accidental al limpiar configuracion persistida en frontend.
- **Cambios principales:**
  - Se agrego dialogo de confirmacion antes de ejecutar `Limpiar sesion local`.
  - Se agrego `SnackBar` de exito tras limpiar storage y resetear estado de pantalla.
  - Se actualizaron README frontend y estado global con esta mejora UX.
- **Archivos clave:**
  - `frontend/lib/pages/home_page.dart`
  - `frontend/README.md`
  - `docs/AGENT_PROJECT_STATUS.md`
- **Validacion:**
  - `flutter analyze`
- **Riesgos/pendientes:**
  - Pendiente internacionalizacion si la app migra a UI bilingue.
- **Siguiente paso recomendado:**
  - Agregar seccion `Ops` en frontend para visualizar `/v1/ops/alerts` y `/metrics` resumido.

## 2026-04-14 - Seccion Ops en frontend local
- **Autor agente:** Codex (Cursor)
- **Contexto:** ampliar validacion operativa desde UI sin depender de llamadas manuales por terminal.
- **Cambios principales:**
  - Se agregaron acciones `Ops Alerts` y `Metrics` en frontend.
  - `Ops Alerts` consume `GET /v1/ops/alerts` con JWT.
  - `Metrics` consume `GET /metrics` y muestra resumen de metricas clave (`http`, `vault_inflight`, `warmup`).
- **Archivos clave:**
  - `frontend/lib/services/nexus_api_client.dart`
  - `frontend/lib/pages/home_page.dart`
  - `frontend/README.md`
- **Validacion:**
  - `flutter analyze`
- **Riesgos/pendientes:**
  - Resumen de metricas usa filtro textual; puede ajustarse si cambian nombres de metricas.
- **Siguiente paso recomendado:**
  - Agregar vista tabular simple para alerts con estado (`triggered`) y severidad.

## 2026-04-14 - Ops alerts visual panel
- **Autor agente:** Codex (Cursor)
- **Contexto:** mejorar legibilidad operativa del frontend sin depender solo de JSON raw.
- **Cambios principales:**
  - Se agrego parse de respuesta `GET /v1/ops/alerts` para renderizar panel visual.
  - Cada alerta muestra nombre, severidad y estado (`ok`/`triggered`) con chips de color.
  - Se mantiene card de JSON crudo para troubleshooting detallado.
- **Archivos clave:**
  - `frontend/lib/pages/home_page.dart`
  - `frontend/README.md`
- **Validacion:**
  - `flutter analyze`
- **Riesgos/pendientes:**
  - Parse depende de formato de respuesta actual (`alerts[]`), requiere ajuste si cambia contrato API.
- **Siguiente paso recomendado:**
  - Extraer modelo `OpsAlert` compartido en `models/` y reutilizarlo en futuras pantallas Ops.
