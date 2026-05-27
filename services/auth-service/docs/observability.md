# Observability Guide

## Overview

The auth-service instruments request execution with OpenTelemetry, emits structured logs with correlation metadata, and records request latency metrics. This supports production debugging, latency analysis, and cross-service traceability.

## OpenTelemetry Setup

- Tracing is initialized at service startup.
- A tracer provider is registered globally.
- HTTP middleware creates spans for inbound requests.
- Context is propagated from middleware through handlers/services/repositories.

## Tracing Flow

1. Request enters tracing middleware.
2. Span starts with route-level operation name.
3. Context carrying span is passed downstream.
4. Latency histogram records elapsed request time.
5. Span closes when response completes.

## Metrics Collected

- HTTP request duration histogram in milliseconds (`http.server.duration_ms`).
- Dimensions can be extended with route/method/status labels as needed.

## Structured Logging

Each request log entry includes:

- method
- path
- status
- request_id
- trace_id
- duration

This enables log-to-trace correlation during incident analysis.

## Correlation IDs and Request IDs

- `request_id` is generated per inbound HTTP request.
- `trace_id` is attached via OTel span context.
- Together they allow:
  - request-level debugging inside a service
  - distributed trace navigation across services

## Latency Tracking

- Middleware records end-to-end HTTP request duration.
- Histogram data can be scraped/exported to dashboards for percentile monitoring (P50/P95/P99).

## Local Visualization Stack (Recommended)

### Grafana

Use Grafana as the observability UI for traces, metrics, and logs.

- Run via Docker Compose profile or standalone Grafana container.
- Connect data sources (Tempo/Jaeger, Prometheus, Loki) depending on environment.

### Tempo or Jaeger

Use one trace backend:

- Tempo for Grafana-native trace storage.
- Jaeger for simpler standalone local tracing workflows.

### How Traces Are Visualized

- Query traces by service name or operation (`GET /api/v1/users/me`, etc.).
- Filter by trace ID captured in logs.
- Inspect span timings to identify hot paths (middleware, DB access, token operations).

### How Logs Correlate With Traces

- Copy `trace_id` from logs.
- Search trace backend for that trace.
- Walk spans to map a failure/latency event to exact code path.

## Screenshot Placeholders

### Tracing

![Login Trace](./screenshots/tracing/login-trace.png)
![Refresh Trace](./screenshots/tracing/refresh-trace.png)
![Me Endpoint Trace](./screenshots/tracing/me-endpoint-trace.png)

### Metrics

![Request Latency Histogram](./screenshots/metrics/request-latency-histogram.png)
![Request Rate Panel](./screenshots/metrics/request-rate.png)

### Dashboards

![Auth Service Dashboard](./screenshots/dashboards/auth-service-dashboard.png)
![Error Rate Dashboard](./screenshots/dashboards/error-rate-dashboard.png)

### Logs

![Structured Request Logs](./screenshots/logs/structured-request-logs.png)
![Trace Correlated Logs](./screenshots/logs/trace-correlated-logs.png)
