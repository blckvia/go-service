package redis

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewClient(logger *zap.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("rdb.host") + ":" + viper.GetString("rdb.port"),
		Password: viper.GetString("rdb.password"),
		DB:       viper.GetInt("rdb.dbname"),
	})

	if err := redisotel.InstrumentTracing(client); err != nil {
		logger.Fatal("failed to instrument redis tracing", zap.Error(err))
		panic(err)
	}

	return client
}
