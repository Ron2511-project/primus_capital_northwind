# Northwind Collect

Herramienta de **cobranza priorizada** para el equipo de finanzas de Northwind (prueba técnica Full-stack Semi-Senior). Responde a: *"¿dónde poner foco hoy?"* — no reemplaza un ERP, sino que ordena la cola de trabajo según segmento, mora y monto.

## Stack

| Capa | Tecnología |
|------|------------|
| API | Go 1.22, arquitectura hexagonal, Chi router |
| DB | PostgreSQL 16 |
| UI | React 18 + Vite + TypeScript |
| Infra | Docker Compose |

## Requisitos

- Docker Desktop (o Docker + Docker Compose v2)
- Opcional para desarrollo local: Go 1.22+, Node 20+

## Levantar en < 10 minutos

```bash
git clone <tu-repo>
cd northwind-collections
cp .env.example .env
docker compose up --build
```

- **Frontend:** http://localhost:5173  
- **API:** http://localhost:8080/health  
- **Swagger UI:** http://localhost:8080/swagger/index.html (requiere internet en el navegador para cargar assets de Swagger)  
- **OpenAPI (YAML):** http://localhost:8080/api/docs/openapi.yaml  

Los datos sintéticos se cargan automáticamente en el primer arranque (`SEED_ON_START=true`).

**Si ves** `FATAL: database "northwind" does not exist` **en los logs de postgres:** era el healthcheck (ya corregido). Reinicia con `docker compose down` y `docker compose up --build`. Si persiste, borra el volumen: `docker compose down -v`.

## Flujo principal (E2E)

1. Abrir el dashboard → ver métricas y **cola priorizada** de clientes en mora.
2. Filtrar por segmento (enterprise / startup / zombie) o prioridad.
3. Clic en un cliente → ver facturas, recomendación y historial.
4. Registrar una acción (llamada, nota, snooze, etc.) → **persiste en PostgreSQL**.
5. Reiniciar contenedores (`docker compose restart`) → los datos y acciones siguen ahí.

## Desarrollo local (sin Docker)

**Base de datos:**

```bash
docker compose up postgres -d
```

**API:**

```bash
cd backend
export DATABASE_URL=postgres://northwind:northwind_secret@localhost:5432/northwind_collections?sslmode=disable
export SEED_ON_START=true
export CORS_ORIGIN=http://localhost:5173
go run ./cmd/api
```

**Frontend:**

```bash
cd frontend
npm install
npm run dev
```

## Estructura del backend (hexagonal)

```
backend/
  cmd/api/                 # entrypoint
  internal/
    domain/                # entidades + reglas (priorización)
    ports/                 # interfaces (driven ports)
    application/           # casos de uso
    adapters/
      http/                # driving adapter (REST)
      postgres/            # driven adapter (persistencia)
      seed/                # datos sintéticos
  migrations/
  docs/openapi.yaml
```

## Documentación adicional

- [DECISIONS.md](./DECISIONS.md) — decisiones de producto y arquitectura
- [AI_LOG.md](./AI_LOG.md) — registro de uso de IA
- [docs/ANALYSIS.md](./docs/ANALYSIS.md) — descomposición del problema
- [docs/database-schema.dbml](./docs/database-schema.dbml) — diagrama ER (importar en dbdiagram.io)

## Variables de entorno

Ver [.env.example](./.env.example).

## API

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/dashboard/summary` | KPIs de cobranza |
| GET | `/api/v1/collection-queue` | Cola priorizada (`?segment=&priority=`) |
| GET | `/api/v1/customers/{id}` | Detalle + facturas + acciones |
| POST | `/api/v1/customers/{id}/actions` | Registrar acción de cobranza |
