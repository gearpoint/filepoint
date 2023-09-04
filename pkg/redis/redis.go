package redis

import (
	"time"

	"github.com/gearpoint/filepoint/config"
	"github.com/go-redis/redis/v8"
)

// Returns new redis client
func NewRedisClient(redisCfg *config.RedisConfig) *redis.Client {
	redisHost := redisCfg.RedisAddr

	if redisHost == "" {
		redisHost = ":6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisHost,
		MinIdleConns: redisCfg.MinIdleConns,
		PoolSize:     redisCfg.PoolSize,
		PoolTimeout:  time.Duration(redisCfg.PoolTimeout) * time.Second,
		Password:     redisCfg.Password, // no password set
		DB:           redisCfg.DB,       // use default DB
	})

	return client
}
