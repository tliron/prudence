package platform

const CACHE_KEY_SEPARATOR = ";"

type CacheKey string

//
// CacheBackend
//

var cacheBackend CacheBackend

type CacheBackend interface {
	LoadRepresentation(key CacheKey) (*CachedRepresentation, bool)  // sync
	StoreRepresentation(key CacheKey, cached *CachedRepresentation) // async
	DeleteRepresentation(key CacheKey)                              // async
	DeleteGroup(name CacheKey)                                      // async
}

func SetCacheBackend(cacheBackend_ CacheBackend) {
	cacheBackend = cacheBackend_
}

func GetCacheBackend() CacheBackend {
	return cacheBackend
}
