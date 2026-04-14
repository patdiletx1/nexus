# Nexus RFP - Agent Handoff Index

Usa este archivo como punto de entrada para cualquier agente nuevo.

## Orden de lectura recomendado (5-10 minutos)
1. `docs/AGENT_HANDOFF_INDEX.md` (este archivo)
2. `docs/AGENT_PROJECT_STATUS.md` (estado real del proyecto y progreso)
3. `docs/AGENT_EXECUTION_PLAYBOOK.md` (como ejecutar trabajo sin romper contexto)
4. `docs/AGENT_CHANGELOG.md` (historial operativo de sesiones)
5. `docs/nexus-sprint-01-backlog.md` (backlog y prioridades)
6. `docs/ops/NXS-009_DASHBOARD_MINIMO.md` (operacion diaria de observabilidad)
7. `docs/ops/NXS-009_ALERT_PLAYBOOK.md` (respuesta ante alertas)
8. `docs/nexus-blueprint-agente.md` (vision completa y arquitectura)
9. `rdf.md` (PRD negocio original)

## Prompt sugerido para arrancar con otro agente
```text
Lee `docs/AGENT_HANDOFF_INDEX.md`, `docs/AGENT_PROJECT_STATUS.md` y `docs/AGENT_EXECUTION_PLAYBOOK.md`.
Luego continua desde la seccion "Siguiente tarea recomendada" en `docs/AGENT_PROJECT_STATUS.md`.
No reescribas arquitectura; implementa, valida con tests, y actualiza estado al cerrar.
```

## Convencion de actualizacion
- Al cerrar una tarea relevante:
  - actualizar `docs/AGENT_PROJECT_STATUS.md` (Hecho / Pendiente / Riesgos)
  - actualizar `docs/nexus-sprint-01-backlog.md` (estado por historia)
- Mantener entradas cortas, concretas y verificables.
