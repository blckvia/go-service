package app

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	h "go-service/internal/handler"
	"go-service/internal/repository"
	"go-service/internal/service"
	n "go-service/pkg/nats"
	p "go-service/pkg/prometheus"
	r "go-service/pkg/redis"
	"go-service/pkg/tracer"
)

// @title Go Service API
// @version 1.0
// @description API Server for Go Service

// @host localhost:8000
// @BasePath /

type App struct {
	Server *http.Server
	Logger *zap.Logger
	Redis  *redis.Client
	Nats   *nats.Conn
	db     *pgxpool.Pool
}

func NewApp(ctx context.Context, logger *zap.Logger) *App {
	redisClient := r.NewClient(logger)

	prometheus.MustRegister(p.CacheHitsTotal)
	prometheus.MustRegister(p.CacheMissesTotal)
	prometheus.MustRegister(p.GoodsCounter)

	if err := InitConfig(); err != nil {
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
	t, err := tracer.InitTracer(viper.GetString("tracer.url"), "go-service")
	if err != nil {
		logger.Fatal("failed to initialize tracer", zap.Error(err))
	}

	nc, err := n.NewNatsQueue(n.Config{
		URL:    viper.GetString("nats.url"),
		Logger: logger,
	})
	if err != nil {
		logger.Fatal("failed to connect to nats", zap.Error(err))
	}
	defer nc.Close()

	redisCache := r.New(redisClient)
	repos := repository.New(ctx, db, redisCache, logger, nc, t)
	services := service.New(repos)
	handlers := h.New(services, t)

	srv := &http.Server{
		Addr:           ":" + viper.GetString("port"),
		Handler:        handlers.InitRoutes(),
		MaxHeaderBytes: 1 << 20, // 1MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	return &App{
		Server: srv,
		Logger: logger,
		Redis:  redisClient,
		Nats:   nc,
		db:     db,
	}
}

// TODO: ping bd and redis etc
func (a *App) Run(ctx context.Context) error {
	_, err := a.Redis.Ping(ctx).Result()
	if err != nil {
		a.Logger.Error("failed to ping redis", zap.Error(err))
	}

	return a.Server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context, logger *zap.Logger) error {
	if a.Nats != nil {
		a.Nats.Close()
	}

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Error("failed to close Redis client", zap.Error(err))
		}
	}

	if a.db != nil {
		a.db.Close()
	}

	return a.Server.Shutdown(ctx)
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
