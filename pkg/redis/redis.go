package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisRepository is a wrapper for Redis.
type RedisRepository struct {
	Client     *redis.Client
	prefix_key string
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
		Client:     client,
		prefix_key: "filepoint",
	}
}

func (r *RedisRepository) getKey(key *string) {
	fmt_key := fmt.Sprintf("%s::%s", r.prefix_key, *key)
	key = &fmt_key
}

func (r *RedisRepository) GetAny(ctx context.Context, key string) ([]byte, error) {
	r.getKey(&key)
	return r.Client.Get(ctx, key).Bytes()
}

func (r *RedisRepository) SetAny(ctx context.Context, key string, value []byte, duration time.Duration) {
	r.getKey(&key)
	err := r.Client.Set(ctx, key, value, duration).Err()
	if err != nil {
		logger.Warn("unable to save request in Redis", zap.Any("key", key), zap.Error(err))
	}
}

func (r *RedisRepository) Del(ctx context.Context, key ...string) {
	for _, k := range key {
		r.getKey(&k)
	}
	r.Client.Del(ctx, key...)
}

func (r *RedisRepository) Exists(ctx context.Context, key string) bool {
	r.getKey(&key)
	return r.Client.Exists(ctx, key).Val() > 0
}

func (r *RedisRepository) GetCachedKeyDuration(ctx context.Context, key string) time.Duration {
	r.getKey(&key)
	return r.Client.TTL(ctx, key).Val()
}
