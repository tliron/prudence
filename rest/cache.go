package rest

//
// CacheBackend
//

var cacheBackend CacheBackend

type CacheBackend interface {
	Load(cacheKey CacheKey) (*CacheEntry, bool)
	Store(cacheKey CacheKey, cacheEntry *CacheEntry)
	Delete(cacheKey CacheKey)
}

// Cache interface
func CacheLoad(context *Context) (*CacheEntry, bool) {
	cacheKey := NewCacheKey(context)
	if cacheEntry, ok := cacheBackend.Load(cacheKey); ok {
		context.Log.Debugf("cache hit: %s|%s", cacheKey, cacheEntry)
		return cacheEntry, true
	} else {
		context.Log.Debugf("cache miss: %s", cacheKey)
		return nil, false
	}
}

// Cache interface
func CacheStore(context *Context) {
	cacheKey := NewCacheKey(context)
	cacheEntry := NewCacheEntry(context)
	cacheBackend.Store(cacheKey, cacheEntry)
	context.Log.Debugf("cache stored: %s|%s", cacheKey, cacheEntry)
}

// Cache interface
func CacheStoreBody(context *Context, encodingType EncodingType, body []byte) {
	cacheKey := NewCacheKey(context)
	cacheEntry := NewCacheEntryBody(context, encodingType, body)
	cacheBackend.Store(cacheKey, cacheEntry)
	context.Log.Debugf("cache stored: %s|%s", cacheKey, cacheEntry)
}
