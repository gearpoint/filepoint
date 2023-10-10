package cache_control

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"go.uber.org/zap"
)

// SignedURLCacheControl is the prefixes cache control type.
type SignedURLCacheControl struct {
	cacheFormat     *views.GetSignedURLResponse
	timeToLive      time.Duration
	redisRepository *redis.RedisRepository
}

// NewSignedURLCacheControl returns a SignedURLCacheControl instance.
func NewSignedURLCacheControl(redisRepository *redis.RedisRepository) *SignedURLCacheControl {
	ttl := aws_repository.SignExpiration - (1 * time.Hour)

	return &SignedURLCacheControl{
		timeToLive:      ttl,
		redisRepository: redisRepository,
	}
}

// Get gets the s3 signed URL response from cache.
func (c *SignedURLCacheControl) Get(ctx context.Context, prefix string) (*views.GetSignedURLResponse, error) {
	cached, err := c.redisRepository.GetAny(ctx, prefix)
	if err != nil {
		return nil, err
	}

	cacheFormat := c.cacheFormat
	err = json.Unmarshal(cached, &cacheFormat)
	if err != nil {
		return nil, err
	}

	return cacheFormat, nil
}

// Add adds the s3 signed URL response to cache.
func (c *SignedURLCacheControl) Add(ctx context.Context, prefix string, cache *views.GetSignedURLResponse) {
	cacheBytes, err := json.Marshal(cache)
	if err == nil {
		c.redisRepository.SetAny(ctx, prefix, cacheBytes, c.timeToLive)
		return
	}

	logger.Warn("unable to set key in Redis", zap.Any("key", prefix), zap.Error(err))
}

// Get gets the s3 signed URL response bytes from cache.
func (c *SignedURLCacheControl) GetBytes(ctx context.Context, prefix string) ([]byte, error) {
	return c.redisRepository.GetAny(ctx, prefix)
}

// Add adds the s3 signed URL response bytes to cache.
func (c *SignedURLCacheControl) AddBytes(ctx context.Context, prefix string, cache []byte) {
	c.redisRepository.SetAny(ctx, prefix, cache, c.timeToLive)
}

// Del deletes one signed URL responses from cache.
func (c *SignedURLCacheControl) Del(ctx context.Context, prefix string) {
	c.redisRepository.Del(ctx, prefix)
}

// DelMany deletes many signed URL responses from cache.
func (c *SignedURLCacheControl) DelMany(ctx context.Context, prefix []string) {
	c.redisRepository.Del(ctx, prefix...)
}
