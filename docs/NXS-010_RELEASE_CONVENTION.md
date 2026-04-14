# NXS-010 - CI/CD and Release Convention

Convencion minima para releases tecnicos iterativos.

## 1) Branching y PR
- Todo cambio entra por Pull Request.
- PR debe pasar workflow `ci` (tests Docker).
- PR usa plantilla de `.github/pull_request_template.md`.

## 2) Formato de release tecnico
- Versionado sugerido: `vMAJOR.MINOR.PATCH`.
- Para este proyecto en etapa temprana:
  - `MINOR`: nuevas capacidades funcionales.
  - `PATCH`: fixes/hardening sin cambios de alcance.

## 3) Changelog tecnico por iteracion
- Registrar cada release en `docs/TECH_CHANGELOG.md`.
- Formato minimo por entrada:
  - version/tag,
  - fecha,
  - cambios,
  - validacion,
  - riesgos residuales.

## 4) Gate de calidad minimo
- `go test ./...` obligatorio (Docker en CI).
- Si se tocan endpoints, actualizar `api/README.md`.
- Si se tocan planes/estado, actualizar docs de handoff.

## 5) Cierre de release
- Confirmar estado limpio (`git status`).
- Crear tag (cuando haya remoto y politica definida).
- Registrar entry en changelog tecnico.
