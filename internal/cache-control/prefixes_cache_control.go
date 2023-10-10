package cache_control

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/utils"
	"go.uber.org/zap"
)

// PrefixesCacheControl is the prefixes cache control type.
type PrefixesCacheControl struct {
	cacheFormat     []string
	timeToLive      time.Duration
	redisRepository *redis.RedisRepository
}

// NewPrefixesCacheControl returns a PrefixesCacheControl instance.
func NewPrefixesCacheControl(redisRepository *redis.RedisRepository) *PrefixesCacheControl {
	return &PrefixesCacheControl{
		cacheFormat:     []string{},
		timeToLive:      12 * time.Hour,
		redisRepository: redisRepository,
	}
}

// Get gets the prefixes from cache.
func (c *PrefixesCacheControl) Get(ctx context.Context, prefixesKey string) ([]string, error) {
	cached, err := c.redisRepository.GetAny(ctx, prefixesKey)
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

// Add adds the prefixes to cache.
func (c *PrefixesCacheControl) Add(ctx context.Context, prefixesKey string, cache []string) {
	cacheBytes, err := json.Marshal(cache)
	if err == nil {
		c.redisRepository.SetAny(ctx, prefixesKey, cacheBytes, c.timeToLive)
		return
	}

	logger.Warn("unable to set key in Redis", zap.Any("key", prefixesKey), zap.Error(err))
}

// AddKeyToCachedPrefixes adds the new key to the cached prefixes.
func (c *PrefixesCacheControl) AddKeyToCachedPrefixes(ctx context.Context, prefix string) {
	if utils.CheckPrefixIsFolder(prefix) {
		return
	}

	prefixesKey := utils.GetPrefixFolder(prefix)
	if !c.redisRepository.Exists(ctx, prefixesKey) {
		return
	}

	cached, err := c.Get(ctx, prefixesKey)
	if err != nil {
		return
	}

	cachePrefixAdded := append(cached, prefix)

	newSliceBytes, err := json.Marshal(cachePrefixAdded)
	if err != nil {
		logger.Warn("unable to set key in Redis", zap.Any("key", prefixesKey), zap.Error(err))
		return
	}

	duration := c.redisRepository.GetCachedKeyDuration(ctx, prefixesKey)
	c.redisRepository.SetAny(ctx, prefixesKey, newSliceBytes, duration)
}

// Del deletes the prefixes from cache.
func (c *PrefixesCacheControl) Del(ctx context.Context, prefixesKey string) {
	c.redisRepository.Del(ctx, prefixesKey)
}
