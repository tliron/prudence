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
	context.Log.Debugf("trying cache: %s", cacheKey)
	return cacheBackend.Load(cacheKey)
}

// Cache interface
func CacheStore(context *Context) {
	cacheKey := NewCacheKey(context)
	cacheEntry := NewCacheEntry(context)
	cacheBackend.Store(cacheKey, cacheEntry)
	context.Log.Debugf("cache stored: %s", cacheKey)
}

// Cache interface
func CacheStoreBody(context *Context, encoding string, body []byte) {
	cacheKey := NewCacheKey(context)
	cacheEntry := NewCacheBody(context, encoding, body)
	cacheBackend.Store(cacheKey, cacheEntry)
	context.Log.Debugf("cache stored: %s", cacheKey)
}
