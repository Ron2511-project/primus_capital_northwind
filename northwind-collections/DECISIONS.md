# Decisiones clave (5–7)

## 1. MVP = “Cola de foco”, no un CRM de cobranza completo

**Decisión:** Construir un dashboard con cola priorizada + registro de acciones, en lugar de automatizar emails, integraciones contables o scoring ML.

**Por qué:** El enunciado pide “saber dónde poner foco” en 3 días, con un equipo de finanzas de 2 personas. El dolor principal es la planilla manual los lunes y los recordatorios genéricos. Una cola accionable entrega valor inmediato y demuestra criterio de producto.

**Descartado:** Motor de emails, pasarelas de pago, roles multi-usuario, reportes históricos avanzados.

---

## 2. Segmentación explícita: enterprise / startup / standard / zombie

**Decisión:** Cada cliente tiene un `segment` que modifica la priorización y la recomendación.

**Por qué:** El caso de negocio distingue explícitamente tres perfiles (grandes con ciclo 75d, startups sin caja, zombis 90+ días). Tratarlos igual explica por qué la mora subió y los recordatorios fallan.

**Supuesto:** El segmento se asigna manualmente (como haría la analista de cobranza hoy); no se infiere automáticamente.

---

## 3. Motor de prioridad en dominio (hexagonal), no en SQL ni en React

**Decisión:** `domain.CalculatePriority()` vive en la capa de dominio; repos solo agregan datos.

**Por qué:** Es la regla de negocio central y debe ser testeable sin HTTP ni DB. En hexagonal, el dominio no depende de adaptadores. Si mañana cambian los criterios, se modifica un solo lugar.

**Trade-off:** La cola se calcula en memoria al listar (~420 clientes es aceptable para el MVP).

---

## 4. Enterprise: no alertar antes de sus términos contractuales

**Decisión:** Clientes `enterprise` con mora menor a `payment_terms_days` (ej. 75) reciben prioridad `monitor` y mensaje “no enviar recordatorio genérico”.

**Por qué:** El documento dice que los grandes pagan a 75 días por proceso interno y que los recordatorios automáticos son ruido para ellos.

---

## 5. Acciones de cobranza como flujo E2E principal

**Decisión:** El POST de acciones (llamada, nota, snooze, etc.) es el flujo que persiste y demuestra el sistema end-to-end.

**Por qué:** Cumple el criterio de aceptación de interacción → persistencia. Reemplaza la columna “¿a quién llamamos?” de la planilla. El `snooze` reduce score 72h para evitar re-contacto inmediato.

**Descartado:** Marcar facturas como pagadas desde UI (se puede agregar en 2 semanas más con webhook bancario).

---

## 6. Datos sintéticos representativos, no anonimización real

**Decisión:** Seed con 12 clientes y facturas que ilustran cada segmento y nivel de mora.

**Por qué:** No hay acceso a datos reales por privacidad. Los montos y plazos reflejan el rango del enunciado (USD 200–15.000 MRR).

---

## 7. Docker Compose de 3 servicios sin orquestación extra

**Decisión:** `postgres` + `api` + `frontend` (nginx sirviendo build estático).

**Por qué:** Cumple “levantar en <10 min” para evaluadores sin instalar Go/Node. Variables vía `.env.example`, sin secretos en código.

**Supuesto:** En producción se usaría migraciones versionadas (golang-migrate) y healthchecks más estrictos; para la prueba, SQL en carpeta `migrations/` es suficiente.
