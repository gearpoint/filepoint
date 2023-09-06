// package redis contains the Redis functions and configurations.
package redis

import (
	"time"

	"github.com/gearpoint/filepoint/config"
	"github.com/go-redis/redis/v8"
)

// NewRedisClient returns a new redis client instance.
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
		Password:     redisCfg.Password,
		DB:           redisCfg.DB,
	})

	return client
}
