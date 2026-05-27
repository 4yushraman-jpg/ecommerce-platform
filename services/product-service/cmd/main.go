package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/4yushraman-jpg/product-service/internal/cache"
	"github.com/4yushraman-jpg/product-service/internal/config"
	"github.com/4yushraman-jpg/product-service/internal/database"
	"github.com/4yushraman-jpg/product-service/internal/handler"
	"github.com/4yushraman-jpg/product-service/internal/middleware"
	"github.com/4yushraman-jpg/product-service/internal/observability"
	"github.com/4yushraman-jpg/product-service/internal/repository"
	"github.com/4yushraman-jpg/product-service/internal/routes"
	"github.com/4yushraman-jpg/product-service/internal/service"
	"github.com/4yushraman-jpg/product-service/internal/validator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := observability.NewLogger()

	shutdownTracer, err := observability.InitTracing(context.Background(), "product-service")
	if err != nil {
		logger.Error("failed to initialize tracing", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdownTracer(context.Background()) }()

	db, err := database.NewPostgres(cfg)
	if err != nil {
		logger.Error("failed to connect postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := cache.NewRedis(cfg)
	if err != nil {
		logger.Error("failed to connect redis", "error", err)
		os.Exit(1)
	}
	defer func() { _ = redisClient.Close() }()

	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(
		productRepo,
		redisClient,
		time.Duration(cfg.CacheTTLSeconds)*time.Second,
		&service.NoopEventPublisher{},
	)
	validate := validator.New()
	productHandler := handler.NewProductHandler(productService, validate)
	healthHandler := handler.NewHealthHandler()

	tracingMiddleware, err := middleware.NewTracingMiddleware()
	if err != nil {
		logger.Error("failed to initialize tracing middleware", "error", err)
		os.Exit(1)
	}

	router := routes.SetupRouter(cfg, logger, tracingMiddleware, healthHandler, productHandler)
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "error", err)
	}

	logger.Info("server exited properly")
}
