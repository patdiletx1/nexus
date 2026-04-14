# Nexus Frontend (Bootstrap)

Frontend Flutter inicial para validar integracion local con backend Nexus.

## Funcionalidad inicial
- Configurar `API_BASE_URL` y `JWT_TOKEN` desde UI.
- Persistencia local automatica de `API_BASE_URL`, `JWT_TOKEN` y `Tender ID`.
- Toggle para mostrar/ocultar `JWT_TOKEN`.
- Boton para limpiar sesion local (resetea storage + campos + respuestas).
- Confirmacion previa y snackbar de resultado al limpiar sesion local.
- Probar endpoints:
  - `GET /health/live`
  - `GET /v1/company/profile`
  - `GET /v1/tenders/sync?limit=20`
  - `GET /v1/tenders?limit=20`
  - `POST /v1/tenders/score/warmup`
  - `GET /v1/tenders/{id}/score`
- Ingresar `Tender ID` para score directo (por defecto `MOCK-003`).

## Ejecutar en local
1. Resolver el bloqueo local de Flutter SDK (`dubious ownership` en `/opt/homebrew/share/flutter`).
2. Desde `frontend/`:
   - `flutter pub get`
   - `flutter run -d chrome` (o dispositivo disponible)

## Backend local recomendado
- `SUPABASE_JWT_SECRET=local-dev-secret`
- `CHILECOMPRA_MOCK_ENABLED=true`
- API en `http://localhost:8080`

## JWT local rapido
Desde raiz del repo:
- `export JWT_TOKEN="$(./scripts/gen_local_jwt.sh)"`

Tambien puedes customizar claims:
- `JWT_COMPANY_ID=company-local-99 JWT_SUB=user-x ./scripts/gen_local_jwt.sh`

## Estructura actual
- `lib/main.dart`: inicializacion de app y tema.
- `lib/pages/home_page.dart`: pantalla principal y estado de UI.
- `lib/services/nexus_api_client.dart`: llamadas HTTP al backend.
- `lib/widgets/response_card.dart`: componente reusable para respuestas.
