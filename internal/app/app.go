package app

import (
	"context"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	n "go-service/pkg/nats"
	r "go-service/pkg/redis"
)

// @title Go Service API
// @version 1.0
// @description API Server for Go Service

// @host localhost:8000
// @BasePath /

// TODO: clickhouse, nats, config application
type App struct {
	Server *http.Server
	Logger *zap.Logger
	Redis  *redis.Client
	Nats   *nats.Conn
}

func NewApp(port string, handler http.Handler, logger *zap.Logger) *App {
	redisClient := r.NewClient(logger)

	nc, err := n.NewNatsQueue(n.Config{
		URL:    viper.GetString("nats.url"),
		Logger: logger,
	})
	if err != nil {
		logger.Fatal("failed to connect to nats", zap.Error(err))
	}
	defer nc.Close()

	return &App{
		Server: &http.Server{
			Addr:           ":" + port,
			Handler:        handler,
			MaxHeaderBytes: 1 << 20, // 1MB
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
		},
		Logger: logger,
		Redis:  redisClient,
		Nats:   nc,
	}
}

func (a *App) Run() error {
	return a.Server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.Server.Shutdown(ctx)
}
