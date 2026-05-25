# Análisis del problema y sub-requerimientos

## Problema (como lo entendí)

Northwind tiene **420 clientes B2B**, facturación mensual ~USD 380k, y un equipo de **2 personas en finanzas** que gestiona cobranza con una **planilla manual**. La mora subió del 6% al 14% sin diagnóstico claro. Los morosos no son homogéneos y los recordatorios automáticos generan ruido u ofenden.

La CEO pide una herramienta para **gestionar cobranza y anticiparse**, con foco en **dónde poner atención** — sin PM ni especificación detallada.

## Sub-requerimientos derivados

| ID | Sub-requerimiento | Incluido en MVP |
|----|-------------------|-----------------|
| R1 | Ver cola de cuentas ordenada por urgencia/impacto | ✅ |
| R2 | Diferenciar tratamiento por segmento de cliente | ✅ |
| R3 | Registrar interacciones de cobranza (persistente) | ✅ |
| R4 | Métricas agregadas (mora, críticos, zombis) | ✅ |
| R5 | Detalle de facturas por cliente | ✅ |
| R6 | Automatización de emails / plantillas | ❌ (fase 2) |
| R7 | Integración ERP / bancos | ❌ (fase 2) |
| R8 | Auth multi-rol | ❌ (supuesto: equipo interno confiable) |
| R9 | ML para predicción de churn por mora | ❌ (reglas explícitas primero) |

## Qué incluí y qué descarté

**Incluido:** Dashboard + cola priorizada + detalle + log de acciones + API documentada + Docker + datos sintéticos.

**Descartado (con razón):** Emails automáticos (el problema dice que fallan), pasarela de pagos (no es el foco de “dónde mirar”), multi-tenant (un solo SaaS vendor).

## Supuestos explícitos (no pregunté a la CEO)

1. **Un solo usuario interno** (“finanzas”) sin login — prioricé el flujo de valor.
2. **Segmento asignado manualmente** al dar de alta el cliente (como categoría en la planilla).
3. **“Zombie”** = segmento + facturas con 90+ días impagas (doble señal).
4. **Moneda USD** como en el enunciado.
5. **No hay API de facturación externa** — facturas en nuestra DB con seed.

## Flujo principal usable

```
Finanzas abre dashboard
  → ve KPIs y cola priorizada
  → filtra por segmento/prioridad
  → abre cliente
  → registra llamada/nota/snooze
  → vuelve al dashboard (prioridad puede bajar si snooze)
```

## Extensión a 2 semanas más

- Reglas de prioridad configurables en UI (sin redeploy).
- Marcar factura pagada + conciliación CSV.
- Integración Slack/email solo para cola `critical`/`high`.
- Auditoría y roles (analista vs jefe).
- Tests de contrato API + e2e Playwright.
