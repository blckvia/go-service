package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"go-service/pkg/redis"
	"go-service/pkg/tracer"

	"go-service/internal/app"
	"go-service/internal/handler"
	"go-service/internal/repository"
	"go-service/internal/service"
)

func main() {
	ctx := context.Background()

	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}(logger)

	if err := initConfig(); err != nil {
		logger.Fatal("error initializing configs: %w", zap.Error(err))
	}

	if err := godotenv.Load(); err != nil {
		logger.Fatal("error loading env variables: %w", zap.Error(err))
	}

	db, err := repository.NewPostgresDB(ctx, repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})

	if err != nil {
		logger.Fatal("failed to initialize db", zap.Error(err))
	}

	_, err = tracer.InitTracer("http://localhost:14268/api/traces", "go-service")
	if err != nil {
		logger.Fatal("failed to initialize tracer", zap.Error(err))
	}

	app := app.NewApp(viper.GetString("port"), nil, logger)
	redisCache := redis.NewRedisCache(app.Redis)
	repos := repository.New(ctx, db, redisCache)
	services := service.New(repos)
	handlers := handler.NewHandler(services)

	app.Server.Handler = handlers.InitRoutes()
	go func() {
		if err := app.Run(); err != nil {
			logger.Fatal("failed to run server: %w", zap.Error(err))
		}
	}()

	logger.Info("app started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info("app shutting down")

	if err := app.Shutdown(context.Background()); err != nil {
		logger.Error("failed to shutdown server: %w", zap.Error(err))
	}

	if err := db.Close(ctx); err != nil {
		logger.Error("failed to close db: %w", zap.Error(err))
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
