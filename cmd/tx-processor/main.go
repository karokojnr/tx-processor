package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	rds "tx-processor/cache/redis"
	"tx-processor/config"
	"tx-processor/db"
	"tx-processor/handlers"
	"tx-processor/logger"
	"tx-processor/repository"
	"tx-processor/server"
	"tx-processor/services"

	"github.com/redis/go-redis/v9"
)

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		return err
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	appLogger := slog.New(jsonHandler)
	loggerWrapper := logger.NewSlogAdapter(appLogger)

	database, err := db.NewPostgresDB(&cfg.DatabaseConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConfig.RedisAddr,
		Password: cfg.RedisConfig.RedisPw,
		DB:       cfg.RedisConfig.RedisDB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis client: %w", err)
	}
	defer redisClient.Close()

	analyticsCache := rds.NewRedisAnalyticsCache(redisClient)
	analyticsRepo := repository.NewAnalyticsRepo(database)

	// Create analytics service
	analyticsService := services.NewAnalyticsService(analyticsRepo, analyticsCache)

	handler := handlers.NewHandler(analyticsService, cfg, loggerWrapper)

	serverCfg := server.Config{
		Port:   cfg.Port,
		Logger: loggerWrapper,
	}

	srv := server.New(serverCfg, handler)
	if err := srv.Start(ctx); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
