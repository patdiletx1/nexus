# NXS-010 - Remote Enforcement (GitHub)

Guia rapida para cerrar NXS-010 en remoto.

## Estado actual
- CI local y workflow listos.
- Falta enforcement en GitHub:
  - required checks en `main`,
  - PR reviews obligatorias,
  - merge directo deshabilitado por politica.

## Paso 1: Autenticacion GitHub CLI
```bash
gh auth login
```

## Paso 2: Crear repo remoto + branch protection
Desde la raiz del proyecto:
```bash
./scripts/setup_github_enforcement.sh nexus private
```

Parametros:
- `nexus`: nombre de repo (opcional, default `nexus`)
- `private|public`: visibilidad (opcional, default `private`)

## Que configura el script
- Crea remoto `origin` si no existe y hace push de `main`.
- Activa protection en `main`:
  - required status check: `backend-go-tests`
  - 1 aprobacion minima de PR
  - dismiss stale reviews
  - linear history
  - conversation resolution obligatoria
  - force push y deletions deshabilitados

## Limitacion conocida (GitHub plan)
En algunos planes, `branch protection` no esta disponible para repos privados.

Si aparece error `403 Upgrade to GitHub Pro...`:
1. Upgrade de plan para habilitar protection en privados, o
2. Cambiar el repo a publico y rerun del script.

## Validacion final
1. Crear PR de prueba.
2. Verificar que merge quede bloqueado hasta pasar `backend-go-tests`.
3. Verificar solicitud de review obligatoria.
