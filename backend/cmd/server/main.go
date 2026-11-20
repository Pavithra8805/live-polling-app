package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"live-polling-app/backend/internal/config"
	"live-polling-app/backend/internal/handlers"
	"live-polling-app/backend/internal/middleware"
	"live-polling-app/backend/internal/repository"
	"live-polling-app/backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("connect to MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			log.Printf("disconnect MongoDB: %v", err)
		}
	}()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	if err := mongoClient.Ping(pingCtx, nil); err != nil {
		cancel()
		log.Fatalf("ping MongoDB: %v", err)
	}
	cancel()

	database := mongoClient.Database(cfg.MongoDatabase)
	userRepository := repository.NewUserRepository(database)
	indexCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := userRepository.EnsureIndexes(indexCtx); err != nil {
		cancel()
		log.Fatalf("create MongoDB indexes: %v", err)
	}
	cancel()

	authService := service.NewAuthService(userRepository, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiry)
	authHandler := handlers.NewAuthHandler(authService)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.SecurityHeaders(), middleware.CORS(cfg.AllowedOrigins), middleware.RequestBodyLimit(cfg.MaxBodyBytes))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/api/auth/signup", authHandler.Signup)
	router.POST("/api/auth/login", authHandler.Login)

	server := &http.Server{Addr: cfg.HTTPAddr, Handler: router}
	go func() {
		log.Printf("server listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
