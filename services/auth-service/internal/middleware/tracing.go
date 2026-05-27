package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type TracingMiddleware struct {
	tracer  trace.Tracer
	latency metric.Float64Histogram
}

func NewTracingMiddleware() (*TracingMiddleware, error) {
	meter := otel.Meter("auth-service")
	latency, err := meter.Float64Histogram("http.server.duration_ms")
	if err != nil {
		return nil, err
	}
	return &TracingMiddleware{
		tracer:  otel.Tracer("auth-service/http"),
		latency: latency,
	}, nil
}

func (m *TracingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		start := time.Now()
		defer func() {
			m.latency.Record(ctx, float64(time.Since(start).Milliseconds()))
			span.End()
		}()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
