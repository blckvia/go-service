package nats

import (
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Config struct {
	URL    string
	Logger *zap.Logger
}

func NewNatsQueue(config Config) (*nats.Conn, error) {
	nc, err := nats.Connect(config.URL)
	if err != nil {
		config.Logger.Error("failed to connect to nats", zap.Error(err))
		return nil, err
	}

	config.Logger.Info("connected to NATS server", zap.String("URL", config.URL))
	return nc, nil
}
