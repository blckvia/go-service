package logger

import "go.uber.org/zap"

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

func GetLogger() *zap.Logger {
	return logger
}
