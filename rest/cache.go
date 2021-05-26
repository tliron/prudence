package rest

import (
	"github.com/tliron/prudence/platform"
)

func CacheLoad(context *Context) (platform.CacheKey, *platform.CacheEntry, bool) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheKey := NewCacheKey(context)
		if cacheEntry, ok := cacheBackend.Load(cacheKey); ok {
			context.Log.Debugf("cache hit: %s|%s", cacheKey, cacheEntry)
			return cacheKey, cacheEntry, true
		} else {
			context.Log.Debugf("cache miss: %s", cacheKey)
			return platform.CacheKey{}, nil, false
		}
	} else {
		return platform.CacheKey{}, nil, false
	}
}

func CacheDelete(context *Context) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheKey := NewCacheKey(context)
		cacheBackend.Delete(cacheKey)
		log.Debugf("cache deleted: %s", cacheKey)
	}
}

func CacheUpdate(cacheKey platform.CacheKey, cacheEntry *platform.CacheEntry) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheBackend.Store(cacheKey, cacheEntry)
		log.Debugf("cache updated: %s|%s", cacheKey, cacheEntry)
	}
}

func CacheStoreContext(context *Context) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheKey := NewCacheKey(context)
		cacheEntry := NewCacheEntryFromContext(context)
		cacheBackend.Store(cacheKey, cacheEntry)
		context.Log.Debugf("cache stored: %s|%s", cacheKey, cacheEntry)
	}
}

func CacheStoreBody(context *Context, encoding platform.EncodingType, body []byte) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheBackend := platform.GetCacheBackend()
		cacheKey := NewCacheKey(context)
		cacheEntry := NewCacheEntryFromBody(context, encoding, body)
		cacheBackend.Store(cacheKey, cacheEntry)
		context.Log.Debugf("cache stored: %s|%s", cacheKey, cacheEntry)
	}
}
