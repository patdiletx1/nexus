# NXS-009 - Alert Playbook (SRE-lite)

Runbook operativo para responder alertas base del backend Nexus.

## Alcance
- Fuente de alertas: `GET /v1/ops/alerts`
- Fuente de metricas: `GET /metrics`
- Alertas cubiertas:
  - `http_error_rate_high`
  - `vault_timeout_rate_high`
  - `vault_inflight_high`
  - `tenders_warmup_skipped_ratio_high`

## Procedimiento general (5 minutos)
1. Confirmar si la alerta sigue activa en 2 lecturas consecutivas (30s de diferencia).
2. Revisar `X-Request-ID` en logs recientes y cruzar endpoint afectado.
3. Determinar impacto:
   - `warning`: degradacion parcial.
   - `critical`: riesgo operacional inmediato.
4. Ejecutar accion de contencion correspondiente.
5. Registrar incidente breve en `docs/AGENT_CHANGELOG.md`.

## Playbook por alerta

### 1) `http_error_rate_high` (warning)
**Significado**
- Aumento de respuestas `5xx` sobre el umbral.

**Checks**
- Endpoints con mayor error: revisar `nexus_http_requests_total`.
- Latencia del endpoint: revisar `nexus_http_request_duration_ms_sum/count`.
- Errores de dependencias externas (ChileCompra, Supabase, Gemini).

**Acciones**
- Si un endpoint puntual falla: aplicar feature degradation temporal.
- Si falla dependencia externa: reducir traffic/rate y activar fallback ya disponible.
- Abrir tarea para fix si persiste > 15 min.

### 2) `vault_timeout_rate_high` (warning)
**Significado**
- El pipeline documental acumula fallos por timeout.

**Checks**
- `nexus_vault_processing_total{error_category="timeout"}`
- Familia documental afectada (`pdf`, `image`, `audio`).
- Estado de proveedor primario y fallback.

**Acciones**
- Priorizar fallback (`simulated`) para mantener continuidad.
- Reintentar items fallidos críticos con `POST /v1/vault/items/{id}/retry`.
- Escalar si timeout > 30 min o afecta mas de una familia documental.

### 3) `vault_inflight_high` (critical)
**Significado**
- Exceso de jobs simultaneos de procesamiento (`nexus_vault_inflight`).

**Checks**
- Volumen de requests a `POST /v1/vault/process`.
- Duracion promedio del pipeline.
- Backlog funcional (items mucho tiempo en `processing`).

**Acciones**
- Limitar disparo de nuevos process requests temporalmente.
- Priorizar reintentos solo para items de alto valor.
- Si persiste > 10 min, tratar como incidente operativo.

### 4) `tenders_warmup_skipped_ratio_high` (warning)
**Significado**
- El warmup de score esta omitiendo demasiados IDs respecto al volumen procesado.

**Checks**
- `nexus_tenders_warmup_skipped_total`
- `nexus_tenders_warmup_processed_total`
- Inputs del cliente (`targeted_ids`, `skipped_ids`) en respuestas recientes.

**Acciones**
- Revisar calidad de `tender_ids` enviados por frontend (duplicados/no existentes).
- Reducir tamano de lote por request si hay overflow de IDs.
- Ajustar fallback a `limit` cuando el modo selectivo venga degradado.

## Cierre de incidente
- Confirmar retorno de alertas a estado no trigger.
- Documentar:
  - causa probable,
  - mitigacion aplicada,
  - follow-up tecnico.
- Actualizar `docs/AGENT_PROJECT_STATUS.md` si cambia riesgo residual.
