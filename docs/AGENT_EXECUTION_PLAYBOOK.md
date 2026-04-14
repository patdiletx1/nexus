# Nexus RFP - Agent Execution Playbook

Guia operativa corta para que cualquier agente continue sin perder contexto.

## 1) Antes de empezar
- Leer en orden:
  - `docs/AGENT_HANDOFF_INDEX.md`
  - `docs/AGENT_PROJECT_STATUS.md`
  - `docs/nexus-sprint-01-backlog.md`
- Confirmar siguiente tarea desde "Siguiente tarea recomendada".

## 2) Regla de ejecucion
- Trabajar en tareas pequenas y cerrables.
- Cada tarea debe terminar con:
  - cambios de codigo,
  - tests ejecutados,
  - estado actualizado en docs.

## 3) Validacion minima obligatoria
- Backend Go:
  - `go test ./...` (preferible por Docker, segun `api/Makefile`)
- Si se tocan migraciones:
  - dejar nombre incremental y descripcion clara.
- Si se tocan endpoints:
  - actualizar `api/README.md`.

## 4) Criterios de calidad
- Mantener aislamiento multi-tenant.
- No romper idempotencia en operaciones async.
- Mantener trazabilidad en auditoria para cambios de estado.
- Preferir fallback seguro cuando falten credenciales externas.

## 5) Cuando cierres una tarea
- Actualizar:
  - `docs/AGENT_PROJECT_STATUS.md` (Hecho/Pendiente/Siguiente)
  - `docs/AGENT_CHANGELOG.md` (entrada de sesion)
  - `docs/nexus-sprint-01-backlog.md` (estado real)
- Dejar nota breve de:
  - que se implemento,
  - como se valido,
  - que queda pendiente.

## 6) Prompt rapido para seguir
```text
Lee `docs/AGENT_HANDOFF_INDEX.md`, `docs/AGENT_PROJECT_STATUS.md` y `docs/AGENT_EXECUTION_PLAYBOOK.md`.
Implementa la siguiente tarea recomendada, ejecuta tests y actualiza estado en docs al terminar.
```
