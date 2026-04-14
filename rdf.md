

# PRD: Nexus RFP – Fábrica Agéntica de Licitaciones

**Versión:** 1.0  
**Estado:** Listo para Diseño y Construcción  
**Target:** Mercado B2B Chile (Construcción, Ingeniería, Servicios Técnicos)

---

## 1. Resumen Ejecutivo
Nexus RFP es un SaaS móvil y web que automatiza el ciclo de vida de las licitaciones públicas (Mercado Público) y privadas en Chile. Utiliza una arquitectura de **Agentes de IA Multimodales** para transformar archivos no estructurados (fotos de terreno, audios, PDFs de licitaciones) en propuestas comerciales y presupuestos técnicos terminados.

---

## 2. Personas de Usuario
* **Patricio (Dueño de PyME/Ingeniero):** Siempre en terreno, no tiene tiempo para leer bases de 100 páginas. Necesita que la IA le avise qué ganar y le redacte el borrador mientras él conduce o supervisa obras.
* **Administrativo Técnico:** Encargado de subir archivos y validar que la IA no cometió errores en los precios unitarios.

---

## 3. Requisitos Funcionales (Core Features)

### 3.1. El Radar de Oportunidades (API ChileCompra)
* Conexión horaria con la API de Mercado Público.
* Filtrado inteligente basado en el "ADN de la empresa" (historial, región, capacidad técnica).
* Notificaciones Push proactivas: *"Licitación encontrada en [Comuna]. Match del 85% con tu historial"*.

### 3.2. La Bóveda (Corporate Vault - RAG)
* **Ingesta Multimodal:** Capacidad de recibir fotos de documentos físicos, hojas de Excel, audios de WhatsApp y PDFs.
* **Categorización Autónoma:** La IA clasifica los documentos (Presupuesto, Certificado, Plano, Factura) sin intervención humana.
* **Indexación Vectorial:** Almacenamiento en Postgres (pgvector) para recuperación semántica de datos históricos.

### 3.3. Generador Agéntico de Propuestas
* **Análisis de Bases:** Lectura de PDFs de licitaciones para extraer requisitos críticos (Fechas, Boletas de Garantía, Anexos Técnicos).
* **Redacción de Borradores:** Generación de archivos PDF/Markdown con la propuesta comercial, adaptada al tono de voz de la empresa.

---

## 4. Arquitectura Técnica y Stack
Para maximizar margen y escalabilidad:

* **Frontend:** Flutter (Mobile-First).
* **Backend:** Go (Golang) – Manejo de concurrencia para agentes.
* **IA Inferencia (Texto/Razonamiento):** **Z.ai** (API) utilizando modelos como Llama 3.3 o DeepSeek-R1 (para lógica compleja).
* **IA Visión (OCR Multimodal):** Gemini 3 Flash (API) para procesar fotos y planos.
* **Base de Datos:** Supabase (Postgres + pgvector + Auth + Storage).
* **Hosting:** Hetzner Cloud (Instancias CPU para el backend).



---

## 5. El Modelo de "Setup Pasivo" (Workflow de Automatización)
El sistema debe configurarse solo durante el "Setup" de $45.000 CLP:
1.  **Trigger:** Pago confirmado vía API (Stripe/Flow).
2.  **Ingesta Inicial:** El usuario carga sus últimos 5-10 presupuestos ganados.
3.  **Agente de Extracción de ADN:** Un worker en Z.ai analiza los archivos, define el "Estilo de Redacción" y las "Reglas de Negocio" (márgenes, proveedores recurrentes).
4.  **Almacenamiento:** El perfil se guarda como un `System Prompt` dinámico asociado al ID del cliente.

---

## 6. Modelo de Datos (Esquema Sugerido)
* **Users/Companies:** Perfil, suscripción, créditos de éxito.
* **Vault_Items:** Metadata del archivo, tipo, hash de almacenamiento, embeddings vectoriales.
* **Tenders:** Datos extraídos de la API de Mercado Público (ID, título, descripción, fecha_cierre).
* **Proposals:** Relación entre Tender y Vault_Items usados para generarla.

---

## 7. KPIs y Éxito del Producto
* **Time-to-Draft:** Una propuesta de 20 páginas debe estar lista en < 2 minutos.
* **Accuracy:** Extracción de montos y fechas con > 95% de precisión (validado con modelos de razonamiento).
* **Churn:** Mantener el uso mensual mediante el Radar (valor continuo).

---

## 8. Instrucciones para el Agente de Construcción
1.  **Fase 1 (Diseño):** Generar el diagrama de secuencia para la ingesta de un audio de WhatsApp que termina como un registro en la DB Vectorial.
2.  **Fase 2 (API):** Definir los endpoints en Go para la integración con Mercado Público y Z.ai.
3.  **Fase 3 (Frontend):** Diseñar una UI minimalista en Flutter centrada en la "Bóveda" y el "Radar".

---