# Product Service

Production-oriented product catalog microservice for a distributed e-commerce platform. The service follows the same layered architecture and transport/persistence separation as `auth-service`.

## Overview

`product-service` owns product CRUD, cursor-based catalog listing, Redis cache-aside reads, and future event-publishing hooks. It is designed to stay small at the handler boundary and keep business logic in the service layer.

## Features

- product CRUD
- validation at the HTTP boundary
- structured JSON response envelopes
- centralized HTTP error mapping
- PostgreSQL persistence with pgx
- Redis cache-aside for detail and list reads
- cursor pagination with category filtering and name search
- OpenTelemetry tracing and request latency metrics
- structured `slog` logging with request and trace correlation
- graceful shutdown with bounded termination
- Kafka-ready event abstraction for future publishing

## Architecture

- Handlers: decode, validate, invoke services, respond
- Services: business rules, cache strategy, and event hooks
- Repositories: SQL access and DB error translation only
- Middleware: request identity, tracing, logging, resilience, and security

```mermaid
flowchart LR
    Client --> Router[Chi Router]
    Router --> Middleware
    Middleware --> Handler
    Handler --> Service
    Service --> Repository
    Repository --> PostgreSQL
    Service --> Redis
    Service --> Publisher["Event Publisher (Kafka-ready abstraction)"]
```

## Cache Design

- Detail key: `product:detail:{id}`
- List key: `product:list:{limit}:{cursor}:{category}:{sort}:{search}`
- Strategy: cache-aside on detail and list reads
- TTL: configured via `CACHE_TTL_SECONDS`
- Invalidation: list cache is cleared on create/update/delete; detail cache is cleared on update/delete

## Pagination

List endpoints use cursor pagination on `(created_at, id)`.

- `limit` defaults to 20 and is capped at 100
- `cursor` is an opaque base64 token
- `sort` supports `asc` and `desc`
- `category` filters by category
- `search` filters by product name

## API

- `POST /api/v1/products`
- `GET /api/v1/products/{id}`
- `GET /api/v1/products`
- `PATCH /api/v1/products/{id}`
- `DELETE /api/v1/products/{id}`

OpenAPI spec: `api/openapi/product.yaml`

## Getting Started

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- PostgreSQL
- Redis
- Goose CLI for migrations

### Environment

Create a `.env` file in `services/product-service` when running locally without Compose:

```env
APP_ENV=development
HTTP_PORT=8081
DB_HOST=localhost
DB_PORT=5434
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=productdb
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
CACHE_TTL_SECONDS=120
REQUEST_TIMEOUT_SECONDS=15
CORS_ALLOWED_ORIGIN=*
```

### Run with Docker Compose

```bash
make docker-up
```

### Run Migrations

```bash
make migrate-up
```

### Run Locally

```bash
make run
```

## Testing

```bash
make test
```

Optional integration tests:

```bash
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5434/productdb?sslmode=disable go test ./internal/repository -run Integration
```

## Observability

- request spans are started by middleware and propagated through handlers, services, and repositories
- DB calls and Redis operations are traced with dedicated spans
- request latency is recorded in a histogram metric
- logs include `request_id` and `trace_id` for correlation

See `docs/observability.md` for the trace/log/metric workflow.

## Redis

- Redis is required for cache-aside reads.
- Connection settings come from `REDIS_ADDR`, `REDIS_PASSWORD`, and `REDIS_DB`.

## Database

The products table is managed by Goose migrations under `internal/database/migrations`.

## Engineering Notes

- pgx-only SQL access, no ORM
- context propagation through every layer
- handlers stay thin
- repositories only access persistence
- service owns business logic and cache policy
- event publishing is abstracted for a later Kafka integration
