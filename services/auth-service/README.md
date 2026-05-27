# Distributed E-Commerce Auth Service

Production-oriented authentication service for a distributed e-commerce platform, focused on secure token lifecycle management, reliable persistence, and operational observability.

## Overview

This service provides identity and session capabilities for upstream APIs and user-facing applications. It is designed as an independent boundary in a distributed architecture, with explicit layering (HTTP -> service -> repository) and context-first request propagation.

Engineering goals:

- predictable and testable auth behavior
- secure token issuance and revocation model
- clean separation of concerns
- strong runtime visibility (traces, metrics, logs)
- deployable local and container workflows

## Features

- JWT-based access authentication
- refresh token persistence and rotation
- user registration, login, logout, token refresh, `me` profile endpoint
- PostgreSQL persistence using pgx (no ORM)
- OpenTelemetry tracing and latency instrumentation
- structured JSON logging (`slog`) with request and trace correlation
- centralized response envelopes and HTTP error mapping
- request validation and domain-level error translation
- security middleware: timeout, recovery, CORS, secure headers, rate limiting
- graceful shutdown with bounded termination window
- Dockerized local environment

## Architecture

The service follows layered architecture:

- **handlers**: HTTP transport concerns (decode/validate/respond)
- **services**: business logic and token lifecycle orchestration
- **repositories**: SQL access and DB error translation
- **middleware**: cross-cutting concerns (security, observability, resilience)

```mermaid
flowchart LR
    Client --> Router[Chi Router]
    Router --> Middleware
    Middleware --> Handler
    Handler --> Service
    Service --> Token[JWT Service]
    Service --> Repository
    Repository --> PostgreSQL
```

```mermaid
sequenceDiagram
    participant C as Client
    participant M as Middleware Chain
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant D as Database

    C->>M: HTTP Request
    M->>H: Enriched Context
    H->>S: Business Operation
    S->>R: Context-aware DB call
    R->>D: SQL via pgx
    D-->>R: Result
    R-->>S: Domain result/error
    S-->>H: Response model
    H-->>C: Standard JSON envelope
```

## Tech Stack

| Category | Technology |
|---|---|
| Language | Go |
| HTTP | `net/http`, Chi |
| Database | PostgreSQL |
| DB Driver | pgx |
| Migrations | Goose |
| Auth | JWT (`github.com/golang-jwt/jwt/v5`) |
| Observability | OpenTelemetry, slog |
| Containers | Docker, Docker Compose |

## Project Structure

```text
auth-service/
  cmd/                    # application bootstrap and dependency wiring
  internal/
    config/               # environment config
    database/             # database bootstrap + migrations
    dto/                  # request/response DTOs
    errors/               # domain errors + HTTP mapping
    handler/              # HTTP handlers
    middleware/           # request middleware pipeline
    model/                # domain models
    observability/        # logger/tracing setup
    repository/           # SQL repositories
    response/             # JSON response helpers
    routes/               # router setup
    service/              # business logic
    token/                # JWT generation/parsing
  api/openapi/            # OpenAPI specification
  docs/                   # architecture and observability docs
```

## Getting Started

### Prerequisites

- Go 1.25+
- Docker + Docker Compose
- PostgreSQL (if running without Docker)
- Goose CLI (for migration commands)

### Environment Setup

Create `.env` in `services/auth-service`:

```env
APP_ENV=development
HTTP_PORT=8080
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=authdb
JWT_SECRET=super-secret-key
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7
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

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_ENV` | no | `development` | runtime environment |
| `HTTP_PORT` | no | `8080` | HTTP server port |
| `DB_HOST` | yes | - | PostgreSQL host |
| `DB_PORT` | yes | - | PostgreSQL port |
| `DB_USER` | yes | - | PostgreSQL user |
| `DB_PASSWORD` | yes | - | PostgreSQL password |
| `DB_NAME` | yes | - | PostgreSQL database |
| `JWT_SECRET` | yes | - | HMAC signing secret for JWT |
| `JWT_ACCESS_TTL_MINUTES` | no | `15` | access token TTL (minutes) |
| `JWT_REFRESH_TTL_DAYS` | no | `7` | refresh token TTL (days) |
| `REQUEST_TIMEOUT_SECONDS` | no | `15` | request context timeout |
| `CORS_ALLOWED_ORIGIN` | no | `*` | allowed CORS origin |

## API Endpoints

| Method | Path | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | register user | no |
| `POST` | `/api/v1/auth/login` | issue access + refresh tokens | no |
| `POST` | `/api/v1/auth/refresh` | rotate refresh token and issue new pair | no |
| `POST` | `/api/v1/auth/logout` | revoke refresh token | no |
| `GET` | `/api/v1/users/me` | current authenticated user profile | bearer JWT |

OpenAPI spec: `api/openapi/auth.yaml`

## Running Tests

```bash
make test
```

Optional integration repository tests (when DB is available):

```bash
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5433/authdb?sslmode=disable go test ./internal/repository -run Integration
```

## Observability

- request-level tracing via OpenTelemetry middleware
- latency metrics histogram for HTTP request durations
- structured logging with `request_id` and `trace_id`
- correlation-friendly logs for trace lookup

See `docs/observability.md` for setup guidance and visualization placeholders.

## Security Features

- bcrypt password hashing
- short-lived access tokens (default 15 minutes)
- refresh token persistence and rotation
- explicit token revocation (logout and rotation)
- rate limiting middleware
- panic recovery and request timeout protection
- secure response headers

## Future Improvements

- Redis-backed distributed rate limiting
- service discovery integration for multi-service deployments
- Kubernetes manifests/Helm chart for production orchestration
- centralized distributed tracing and metrics aggregation stack
- event-driven auth workflows (audit events, token revocation events)

## Resume-Grade Engineering Highlights

- clean layered architecture with explicit dependency injection
- production middleware pipeline for resilience and security
- token lifecycle management with rotation and revocation guarantees
- observability-first design with trace/log correlation
- robust error boundaries between domain, transport, and persistence
