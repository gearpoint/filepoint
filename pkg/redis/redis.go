package redis

import (
	"context"
	"time"

	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisRepository is a wrapper for Redis.
type RedisRepository struct {
	Client *redis.Client
}

// NewRedisRepository returns a new redis client.
func NewRedisRepository(redisConfig *config.RedisConfig) *RedisRepository {
	client := redis.NewClient(&redis.Options{
		Addr:         redisConfig.Addr,
		MinIdleConns: redisConfig.MinIdleConns,
		PoolSize:     redisConfig.PoolSize,
		PoolTimeout:  time.Duration(redisConfig.PoolTimeout) * time.Second,
		Username:     redisConfig.Username,
		Password:     redisConfig.Password,
	})

	return &RedisRepository{
		Client: client,
	}
}

func (r *RedisRepository) GetAny(ctx context.Context, key string) ([]byte, error) {
	return r.Client.Get(ctx, key).Bytes()
}

func (r *RedisRepository) SetAny(ctx context.Context, key string, value []byte, duration time.Duration) {
	err := r.Client.Set(ctx, key, value, duration).Err()
	if err != nil {
		logger.Warn("unable to save request in Redis", zap.Any("key", key), zap.Error(err))
	}
}

func (r *RedisRepository) Del(ctx context.Context, key ...string) {
	r.Client.Del(ctx, key...)
}

func (r *RedisRepository) Exists(ctx context.Context, key string) bool {
	return r.Client.Exists(ctx, key).Val() > 0
}

func (r *RedisRepository) GetCachedKeyDuration(ctx context.Context, key string) time.Duration {
	return r.Client.TTL(ctx, key).Val()
}
