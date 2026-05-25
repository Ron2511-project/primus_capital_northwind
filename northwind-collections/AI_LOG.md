# AI_LOG — Registro de uso de Cursor / Agente IA

## Prompt 1 (inicial)

> Necesito dar solución a este challenge, usando Go arquitectura hexagonal como API y frontend en React, dockerizar, explicar solución y razonamiento, documentar supuestos.

**Qué delegué:** Estructura completa del repo, boilerplate hexagonal, Docker, seed, UI funcional, documentación DECISIONS/README.

**Qué no delegué (criterio humano):** Definición del MVP (“cola de foco” vs CRM), reglas de prioridad por segmento, y qué features dejar fuera del alcance 3 días.

**Dónde ayudó:** Velocidad para cumplir criterios mínimos (OpenAPI, .env.example, estados loading/error, dbml) sin perder tiempo en scaffolding.

**Corrección manual recomendada:** Revisar que los mensajes de `recommended_action` suenen naturales para el equipo de finanzas chileno; ajustar copy en demo.

---

## Prompt 2 (implícito — arquitectura)

**Decisión:** Paquete HTTP renombrado internamente a `httpadapter` para evitar conflicto con `net/http` de Go.

**Reflexión:** La IA tiende a crear carpetas llamadas `http`; conviene nombrar `adapters/inbound/http` desde el inicio.

---

## Qué decidí NO delegar en una entrega real

- Validar supuestos con la analista de cobranza (30 min del enunciado).
- Preguntar a jrain@primuscapital.cl si el segmento zombie es etiqueta manual o regla (90d sin pago).
- Escribir 2–3 tests de `CalculatePriority` para enterprise vs startup (la prueba valora “tests correctos, no muchos”).
- Historial de commits granulares hechos por el candidato, no un solo squash de la IA.

---

## Honestidad para la evaluación

Este proyecto fue generado asistido por IA en una sesión. Para la presentación de 45 min, el candidato debe poder explicar cada decisión en `DECISIONS.md`, demo el flujo E2E, y discutir extensiones a 2 semanas (pagos, integración contable, reglas configurables).
