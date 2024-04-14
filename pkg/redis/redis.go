package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func NewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("rdb.host") + ":" + viper.GetString("rdb.port"),
		Password: viper.GetString("rdb.password"),
		DB:       viper.GetInt("rdb.dbname"),
	})

	return client
}
