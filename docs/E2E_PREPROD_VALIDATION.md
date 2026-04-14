# E2E Pre-Production Validation

Checklist operativo para validar integraciones reales antes de pasar a uso productivo.

## Prerequisitos
- API desplegada y accesible (`API_BASE_URL`).
- JWT valido de un usuario con `company_id` real.
- Variables reales cargadas en backend:
  - `DATABASE_URL`
  - `SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY`, `SUPABASE_STORAGE_BUCKET`
  - `GEMINI_API_KEY`
  - `CHILECOMPRA_BASE_URL`, `CHILECOMPRA_API_KEY`

## Script de smoke E2E
Desde raiz del repo:
```bash
export API_BASE_URL="https://<tu-api>"
export JWT_TOKEN="<bearer-jwt>"
export RUN_SYNC=1
export SYNC_LIMIT=20
export WARMUP_LIMIT=50

./api/scripts/e2e_preprod_smoke.sh
```

## Que valida el smoke
1. Health endpoints (`/health/live`, `/health/ready`)
2. Acceso autenticado a `company/profile`
3. Sync real de licitaciones (`/v1/tenders/sync`)
4. Warmup de score cache (`/v1/tenders/score/warmup`)
5. Score de una licitacion real (`/v1/tenders/{id}/score`)

## Criterios de salida recomendados
- Sin errores `5xx` en smoke.
- `synced_count > 0` en al menos 1 corrida.
- `processed_count > 0` en warmup.
- Score responde con `reasons` no vacio.
- Alertas operativas (`/v1/ops/alerts`) sin `critical` persistente.

## Evidencia a guardar
- Salida completa del script (timestamped).
- Captura de `/metrics` y `/v1/ops/alerts`.
- Nota de riesgos residuales en `docs/AGENT_CHANGELOG.md`.
