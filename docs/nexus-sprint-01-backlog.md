# Nexus RFP - Backlog Tecnico Operativo

## Estado actual (resumen ejecutivo)
- **Sprint 1:** mayormente completado en backend/core.
- **Riesgo principal pendiente:** falta validacion en entorno Supabase real (DB + Storage + llaves reales).
- **Siguiente foco:** Radar ChileCompra, hardening OCR productivo y cierre operativo (SRE/CI/CD).

### Avance por historia (Sprint 1)
- **NXS-001 (Schema + RLS):** `done`
- **NXS-002 (API base + auth + health):** `done`
- **NXS-003 (Boveda upload + estados):** `done` (con signer real/fallback)
- **NXS-004 (Pipeline inicial):** `done` (async + extractor simulado/Gemini)
- **NXS-005 (Observabilidad minima):** `in_progress` (logs y trazabilidad listos, faltan metricas/alertas formales)

## Objetivo del sprint
Dejar operativo el esqueleto de plataforma para comenzar el flujo real de:
- autenticacion multiempresa,
- base de datos con aislamiento por tenant,
- API Go base,
- carga y procesamiento inicial de documentos en Boveda.

Duracion sugerida: 2 semanas.

## Prioridades del sprint
1. Seguridad y aislamiento de datos.
2. Base de API estable.
3. Ingesta documental minima funcional.
4. Observabilidad y evidencia de funcionamiento.

## Historias de usuario y tareas tecnicas

### NXS-001 - Configurar esquema Supabase base con RLS
**Tipo:** Backend/Datos  
**Prioridad:** Alta  
**Dependencias:** Ninguna

**Descripcion**
Como plataforma multiempresa, necesitamos tablas base y politicas RLS para evitar filtracion de datos entre clientes.

**Alcance tecnico**
- Crear tablas iniciales: `companies`, `users`, `vault_items`, `audit_events`.
- Definir claves foraneas por `company_id`.
- Crear politicas RLS por tenant y por rol basico.
- Dejar migraciones versionadas.

**Criterios de aceptacion**
- Un usuario de `company_a` no puede leer/escribir datos de `company_b`.
- Todas las tablas multiempresa tienen RLS activo.
- Migraciones aplican desde cero sin intervencion manual.

**Evidencia requerida**
- SQL/migraciones en repo.
- Resultado de pruebas de aislamiento.
- Captura o log de migracion ejecutada.

---

### NXS-002 - Levantar API Go con autenticacion y health checks
**Tipo:** Backend  
**Prioridad:** Alta  
**Dependencias:** NXS-001

**Descripcion**
Como base de desarrollo, necesitamos una API con estructura limpia, autenticacion y endpoints de salud.

**Alcance tecnico**
- Estructura de proyecto Go por capas (handlers, services, repositories).
- Middleware de autenticacion (token de Supabase o estrategia equivalente).
- Endpoints: `GET /health/live`, `GET /health/ready`.
- Manejo de errores estandar con codigos y mensajes consistentes.

**Criterios de aceptacion**
- API inicia en entorno local con configuracion por variables de entorno.
- Endpoints de salud responden correctamente.
- Endpoints protegidos rechazan requests sin autenticacion valida.

**Evidencia requerida**
- Logs de arranque.
- Pruebas unitarias de middleware auth.
- Ejemplos de respuestas HTTP (200/401/500).

---

### NXS-003 - Implementar carga de documentos a Boveda
**Tipo:** Backend/Storage  
**Prioridad:** Alta  
**Dependencias:** NXS-001, NXS-002

**Descripcion**
Como usuario tecnico, quiero subir documentos para que la plataforma pueda procesarlos y reutilizarlos en propuestas.

**Alcance tecnico**
- Endpoint `POST /v1/vault/upload` para emitir URL firmada.
- Guardado de metadata en `vault_items`.
- Estados iniciales: `uploaded`, `processing`, `processed`, `failed`.
- Validaciones de tipo y tamano de archivo.

**Criterios de aceptacion**
- Usuario autenticado puede subir PDF/imagen/audio.
- Se guarda registro en DB con `company_id` correcto.
- Estado inicial se visualiza y cambia durante el proceso.

**Evidencia requerida**
- Request/response de upload.
- Registro creado en DB.
- Caso de error por archivo invalido.

---

### NXS-004 - Pipeline inicial de procesamiento (OCR + clasificacion minima)
**Tipo:** IA/Backend  
**Prioridad:** Media-Alta  
**Dependencias:** NXS-003

**Descripcion**
Como sistema, necesito procesar archivos subidos para extraer texto y clasificar tipo documental de forma preliminar.

**Alcance tecnico**
- Worker asincrono para tomar `vault_items` en estado `uploaded`.
- OCR inicial con proveedor configurado.
- Clasificacion documental basica (presupuesto, certificado, plano, factura, otro).
- Guardar resultado de texto y metadata estructurada.

**Criterios de aceptacion**
- Un documento subido termina en `processed` o `failed` con motivo.
- El texto extraido se guarda y es trazable al archivo fuente.
- La clasificacion se persiste como metadata.

**Evidencia requerida**
- Log de ejecucion del worker.
- Ejemplo de documento procesado exitosamente.
- Ejemplo de error controlado con reintento.

---

### NXS-005 - Observabilidad minima de plataforma
**Tipo:** Plataforma/SRE-lite  
**Prioridad:** Media  
**Dependencias:** NXS-002, NXS-004

**Descripcion**
Como equipo, necesitamos visibilidad de errores y latencia para operar el sistema con confianza.

**Alcance tecnico**
- Logs estructurados (`trace_id`, `company_id`, `endpoint`, `agent_name`).
- Metricas base por endpoint (latencia, tasa de error).
- Dashboard minimo local o en herramienta de monitoreo definida.

**Criterios de aceptacion**
- Cada request genera trazabilidad correlacionable.
- Se puede identificar rapidamente una falla de procesamiento.
- Existen alertas iniciales para error rate alto.

**Evidencia requerida**
- Muestra de logs estructurados.
- Captura de metrica de latencia/error.
- Documento breve de troubleshooting.

## Orden de ejecucion recomendado
1. NXS-001
2. NXS-002
3. NXS-003
4. NXS-004
5. NXS-005

## Definicion de terminado del sprint
- Historias NXS-001 a NXS-004 completadas.
- NXS-005 al menos en version minima operativa.
- Demo de flujo completo: login -> upload -> procesamiento -> estado final.
- Evidencia consolidada en un solo reporte del sprint.

## Riesgos del sprint y mitigaciones
- **Riesgo:** retraso en autenticacion/RLS.  
  **Mitigacion:** bloquear nuevas features hasta cerrar NXS-001/NXS-002.
- **Riesgo:** OCR inestable por calidad documental.  
  **Mitigacion:** registrar confidence score y fallback manual.
- **Riesgo:** falta de trazabilidad en errores.  
  **Mitigacion:** no cerrar historias sin logs estructurados y trace_id.

## Checklist de cierre por historia
- [ ] Codigo implementado.
- [ ] Tests minimos ejecutados.
- [ ] Criterios de aceptacion validados.
- [ ] Evidencia adjunta.
- [ ] Riesgos residuales documentados.

---

## Sprint 2 (proximo - 2 semanas)

### Objetivo
Cerrar brechas de produccion temprana: Radar ChileCompra, robustez de procesamiento documental y operacion minima para despliegue controlado.

### Estado de trabajo
- `done`: terminado con evidencia.
- `in_progress`: en ejecucion activa.
- `next`: listo para tomar.
- `blocked`: depende de acceso/credenciales/proveedor externo.

### Historias Sprint 2

#### NXS-006 - Integracion real ChileCompra (Radar base)
**Estado:** `done` (pendiente validacion final con credenciales reales)  
**Prioridad:** Alta  
**Dependencias:** NXS-002

**Alcance**
- Cliente API ChileCompra con sync incremental.
- Endpoint de ingesta/sync para oportunidades.
- Persistencia de licitaciones en tabla `tenders` (nueva migracion).

**Definition of Done**
- Sync manual y programable funcionando.
- Al menos 1 flujo de consulta en API con filtros basicos.
- Manejo de errores/cuotas del proveedor documentado.

#### NXS-007 - Score inicial de match empresa-licitacion
**Estado:** `done`  
**Prioridad:** Alta  
**Dependencias:** NXS-006

**Alcance**
- Scoring simple por rubro, region, historico documental.
- Endpoint de detalle de score por licitacion.

**Definition of Done**
- Score explicable (motivos visibles).
- Test de regresion de scoring base.

#### NXS-008 - Hardening de procesamiento documental
**Estado:** `in_progress` (clasificacion de errores + dataset mini de doc_type y cobertura de campos clave implementados)  
**Prioridad:** Alta  
**Dependencias:** NXS-003, NXS-004

**Alcance**
- Ajuste por tipo documental (pdf/imagen/audio).
- Mejor manejo de errores por extractor.
- Dataset mini de validacion y umbrales de calidad iniciales.

**Definition of Done**
- Reporte de precision inicial en campos clave.
- Errores clasificados por tipo y accion de fallback.

#### NXS-009 - Observabilidad SRE-lite (metricas + alertas)
**Estado:** `next`  
**Prioridad:** Media-Alta  
**Dependencias:** NXS-005

**Alcance**
- Metricas de latencia/error por endpoint y pipeline.
- Alertas basicas (error rate, cola/timeout procesamiento).
- Dashboard operativo minimo.

**Definition of Done**
- Al menos 3 alertas activas y testeadas.
- Dashboard compartible para monitoreo diario.

#### NXS-010 - CI/CD y gate de calidad
**Estado:** `next`  
**Prioridad:** Media  
**Dependencias:** NXS-002

**Alcance**
- Pipeline con test en Docker y checks obligatorios.
- Convencion de release + changelog tecnico por iteracion.

**Definition of Done**
- PR sin tests no puede mergear.
- Build reproducible en entorno limpio.

### Bloqueadores actuales
- Credenciales reales de Supabase y Gemini para pruebas E2E completas.
- Confirmacion de estrategia final de consumo ChileCompra (limites/cuotas/retries).

### Plan de ejecucion recomendado (proximo)
1. NXS-008
2. NXS-009
3. NXS-010
