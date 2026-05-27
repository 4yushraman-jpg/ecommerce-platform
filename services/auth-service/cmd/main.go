package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/4yushraman-jpg/auth-service/internal/config"
	"github.com/4yushraman-jpg/auth-service/internal/database"
	"github.com/4yushraman-jpg/auth-service/internal/handler"
	"github.com/4yushraman-jpg/auth-service/internal/middleware"
	"github.com/4yushraman-jpg/auth-service/internal/observability"
	"github.com/4yushraman-jpg/auth-service/internal/repository"
	"github.com/4yushraman-jpg/auth-service/internal/routes"
	"github.com/4yushraman-jpg/auth-service/internal/service"
	"github.com/4yushraman-jpg/auth-service/internal/token"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("failed to load .env file")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := observability.NewLogger()
	shutdownTracer, err := observability.InitTracing(context.Background(), "auth-service")
	if err != nil {
		logger.Error("failed to initialize tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			logger.Error("failed to shutdown tracer", "error", err)
		}
	}()

	db, err := database.NewPostgres(cfg)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	healthHandler := handler.NewHealthHandler()

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	passwordService := service.NewPasswordService()
	jwtService := token.NewJWTService(
		cfg.JWTSecret,
		time.Duration(cfg.JWTAccessTTLMinutes)*time.Minute,
		time.Duration(cfg.JWTRefreshTTLDays)*24*time.Hour,
	)
	tracingMiddleware, err := middleware.NewTracingMiddleware()
	if err != nil {
		logger.Error("failed to initialize tracing middleware", "error", err)
		os.Exit(1)
	}

	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		passwordService,
		jwtService,
		time.Duration(cfg.JWTRefreshTTLDays)*24*time.Hour,
	)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo)

	router := routes.SetupRouter(
		cfg,
		logger,
		jwtService,
		tracingMiddleware,
		healthHandler,
		authHandler,
		userHandler,
	)

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", "port", cfg.HTTPPort)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "error", err)
	}

	logger.Info("server exited properly")
}
