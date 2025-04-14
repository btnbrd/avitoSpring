package main

import (
	"avitoSpring/internal/config"
	"avitoSpring/internal/handlers"
	"avitoSpring/internal/middleware"
	"avitoSpring/internal/server"
	"avitoSpring/internal/services"
	"avitoSpring/internal/storage/pg"
	"context"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := middleware.InitLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	store, err := pg.NewStorage(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer store.DB.Close()

	jwtService := services.NewJWTService()
	authService := services.NewAuthService(store.UserStorage, jwtService)
	productService := services.NewProductService(store.ProductStorage, store.ReceptionStorage)
	pvzService := services.NewPVZService(store.PVZStorage)
	receiptService := services.NewReceptionService(store.ReceptionStorage)

	s := server.NewServer(cfg)

	s.Use(middleware.Logging(logger))

	handlers.RegisterHandlers(s, authService, pvzService, receiptService, productService, jwtService)

	if err := s.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
