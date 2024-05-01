package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"go-service/internal/app"
)

func main() {
	ctx := context.Background()

	if err := app.InitConfig(); err != nil {
		panic(err)
	}

	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}(logger)

	app := app.NewApp(ctx, logger)

	go func() {
		if err := app.Run(ctx); err != nil {
			app.Logger.Fatal("failed to run server: %w", zap.Error(err))
		}
	}()

	app.Logger.Info("app started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	app.Logger.Info("app shutting down")

	if err := app.Shutdown(context.Background(), logger); err != nil {
		app.Logger.Error("failed to shutdown server: %w", zap.Error(err))
	}
}
