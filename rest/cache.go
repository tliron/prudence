package rest

import "github.com/tliron/prudence/platform"

func CacheLoad(context *Context) (*platform.CacheEntry, bool) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheKey := NewCacheKey(context)
		if cacheEntry, ok := cacheBackend.Load(cacheKey); ok {
			context.Log.Debugf("cache hit: %s|%s", cacheKey, cacheEntry)
			return cacheEntry, true
		} else {
			context.Log.Debugf("cache miss: %s", cacheKey)
			return nil, false
		}
	} else {
		return nil, false
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
