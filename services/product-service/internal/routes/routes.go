package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/4yushraman-jpg/product-service/internal/config"
	"github.com/4yushraman-jpg/product-service/internal/handler"
	"github.com/4yushraman-jpg/product-service/internal/middleware"
)

func SetupRouter(
	cfg *config.Config,
	logger *slog.Logger,
	tracingMiddleware *middleware.TracingMiddleware,
	healthHandler *handler.HealthHandler,
	productHandler *handler.ProductHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS(cfg.CORSAllowedOrigin))
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.RequestID)
	r.Use(middleware.RateLimit(150, time.Minute))
	r.Use(middleware.Timeout(time.Duration(cfg.RequestTimeoutSeconds) * time.Second))
	r.Use(tracingMiddleware.Handler)
	r.Use(middleware.Logging(logger))

	r.Get("/health", healthHandler.Health)
	r.Route("/api/v1/products", func(r chi.Router) {
		r.Post("/", productHandler.Create)
		r.Get("/", productHandler.List)
		r.Get("/{id}", productHandler.GetByID)
		r.Patch("/{id}", productHandler.Update)
		r.Delete("/{id}", productHandler.Delete)
	})
	return r
}
