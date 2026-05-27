# Observability Guide

## Overview

`product-service` is instrumented for trace/log correlation and latency visibility. The current setup mirrors the auth-service pattern while adding Redis and product-catalog specific spans.

## OpenTelemetry Setup

- tracing is initialized at application startup
- a global tracer provider is registered
- the HTTP middleware creates a request span for each inbound call
- context is propagated through handlers, services, repositories, and cache calls

## Tracing Flow

1. Request enters the Chi router.
2. Middleware creates a request span and records latency.
3. Handler validates input and calls the service.
4. Service creates spans around cache and business operations.
5. Repository creates spans around PostgreSQL work.
6. The span tree closes when the response completes.

## Metrics

- HTTP request duration histogram: `http.server.duration_ms`
- latency is captured end-to-end at the middleware boundary

## Structured Logging

Request log entries include:

- method
- path
- status
- request_id
- trace_id
- duration

This keeps log-to-trace correlation straightforward when debugging production failures.

## Correlation IDs

- `request_id` is generated per request by middleware
- `trace_id` comes from the active OTel span
- both values are written to the request log entry

## Redis and DB Visibility

- Redis cache gets, sets, deletes, and invalidations are traced with dedicated spans
- PostgreSQL repository calls are traced with per-operation spans
- cache misses are treated as normal flow and should not surface as errors

## Local Workflow

- use Docker Compose for a full local stack
- inspect spans via the configured stdout exporter
- inspect request logs directly from the service output