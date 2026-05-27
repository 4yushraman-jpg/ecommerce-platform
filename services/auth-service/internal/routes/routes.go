package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/4yushraman-jpg/auth-service/internal/config"
	"github.com/4yushraman-jpg/auth-service/internal/handler"
	"github.com/4yushraman-jpg/auth-service/internal/middleware"
	"github.com/4yushraman-jpg/auth-service/internal/token"
)

func SetupRouter(
	cfg *config.Config,
	logger *slog.Logger,
	jwtService *token.Service,
	tracingMiddleware *middleware.TracingMiddleware,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS(cfg.CORSAllowedOrigin))
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.RequestID)
	r.Use(middleware.RateLimit(100, time.Minute))
	r.Use(middleware.Timeout(time.Duration(cfg.RequestTimeoutSeconds) * time.Second))
	r.Use(tracingMiddleware.Handler)
	r.Use(middleware.Logging(logger))

	r.Get("/health", healthHandler.Health)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	r.Route("/api/v1/users", func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtService))
		r.Get("/me", userHandler.Me)
	})

	return r
}
