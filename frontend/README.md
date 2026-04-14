# Nexus Frontend (Bootstrap)

Frontend Flutter inicial para validar integracion local con backend Nexus.

## Funcionalidad inicial
- Configurar `API_BASE_URL` y `JWT_TOKEN` desde UI.
- Probar endpoints:
  - `GET /health/live`
  - `GET /v1/company/profile`
  - `GET /v1/tenders?limit=20`

## Ejecutar en local
1. Resolver el bloqueo local de Flutter SDK (`dubious ownership` en `/opt/homebrew/share/flutter`).
2. Desde `frontend/`:
   - `flutter pub get`
   - `flutter run -d chrome` (o dispositivo disponible)

## Backend local recomendado
- `SUPABASE_JWT_SECRET=local-dev-secret`
- `CHILECOMPRA_MOCK_ENABLED=true`
- API en `http://localhost:8080`
