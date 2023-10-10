package cache_control

import (
	"context"

	"github.com/gearpoint/filepoint/pkg/logger"
	"github.com/gearpoint/filepoint/pkg/redis"
	"github.com/gearpoint/filepoint/pkg/utils"
	"go.uber.org/zap"
)

// UploadCacheControl is the prefixes cache control type.
type UploadCacheControl struct {
	PrefixesCacheControl  *PrefixesCacheControl
	SignedURLCacheControl *SignedURLCacheControl
}

// NewUploadCacheControl returns an UploadCacheControl instance.
func NewUploadCacheControl(redisRepository *redis.RedisRepository) *UploadCacheControl {
	return &UploadCacheControl{
		PrefixesCacheControl:  NewPrefixesCacheControl(redisRepository),
		SignedURLCacheControl: NewSignedURLCacheControl(redisRepository),
	}
}

// RemoveFolderFromCache removes the folder and its objects from cache.
func (c *UploadCacheControl) RemoveFolderFromCache(ctx context.Context, prefixesKey string, prefixes []string) {
	if prefixes != nil {
		c.SignedURLCacheControl.DelMany(ctx, prefixes)
	}
	logger.Debug("asd", zap.Any("asd", prefixesKey))

	c.PrefixesCacheControl.Del(ctx, prefixesKey)
}

// RemoveKeyFromCachedPrefixes removes a s3 key from the cached prefixes.
func (c *UploadCacheControl) RemoveKeyFromCachedPrefixes(ctx context.Context, prefix string) {
	c.SignedURLCacheControl.Del(ctx, prefix)

	prefixesKey := utils.GetPrefixFolder(prefix)
	if prefixesKey == "" {
		return
	}

	prefixes, err := c.PrefixesCacheControl.Get(ctx, prefixesKey)
	if err != nil {
		return
	}

	newPrefixesCache := []string{}
	for _, value := range prefixes {
		if value == prefix {
			continue
		}
		newPrefixesCache = append(newPrefixesCache, value)
	}

	c.PrefixesCacheControl.Add(ctx, prefixesKey, newPrefixesCache)
}
