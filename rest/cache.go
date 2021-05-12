package rest

import (
	"sync"
	"time"
)

// TODO goroutine to prune cache

var cache sync.Map

func ToCache(context *Context) {
	cacheKey := NewCacheKey(context)
	cacheEntry := NewCacheEntry(context)
	cache.Store(cacheKey, cacheEntry)
	context.Log.Debugf("cache store: %s, until %s", cacheKey.Path, cacheEntry.Expiration)
}

func FromCache(context *Context) (*CacheEntry, bool) {
	cacheKey := NewCacheKey(context)
	if cacheEntry, ok := cache.Load(cacheKey); ok {
		cacheEntry_ := cacheEntry.(*CacheEntry)
		if cacheEntry_.Expired() {
			context.Log.Debugf("cache expired: %s", cacheKey.Path)
			cache.Delete(cacheKey)
			return nil, false
		} else {
			context.Log.Debugf("cache hit: %s", cacheKey.Path)
			return cacheEntry_, true
		}
	} else {
		return nil, false
	}
}

func PruneCache() {
	cache.Range(func(key interface{}, value interface{}) bool {
		if value.(*CacheEntry).Expired() {
			cache.Delete(key)
		}
		return true
	})
}

//
// CacheKey
//

type CacheKey struct {
	Path        string
	ContentType string
}

func NewCacheKey(context *Context) CacheKey {
	return CacheKey{
		Path:        context.Path,
		ContentType: context.ContentType,
	}
}

//
// CacheEntry
//

type CacheEntry struct {
	Headers      [][][]byte
	Body         []byte
	LastModified time.Time
	Expiration   time.Time
	// TODO: for different encodings?
}

func NewCacheEntry(context *Context) *CacheEntry {
	var body []byte
	if context.Context.Request.Header.IsGet() {
		// Body exists only in GET
		body = copyBytes(context.Context.Response.Body())
	}

	var headers [][][]byte
	context.Context.Response.Header.VisitAll(func(key []byte, value []byte) {
		// This is an annoying way to get all headers, but unfortunately if we
		// get the entire header via Header() there is no API to set it correctly
		// in CacheEntry.Write
		headers = append(headers, [][]byte{copyBytes(key), copyBytes(value)})
	})

	return &CacheEntry{
		Body:         body,
		Headers:      headers,
		LastModified: context.LastModified,
		Expiration:   time.Now().Add(time.Duration(context.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func (self *CacheEntry) Write(context *Context) {
	context.Context.Response.Reset()

	// Headers
	context.Context.Response.Header.DisableNormalizing() // annoyingly this was re-enabled by Reset above
	for _, header := range self.Headers {
		context.Context.Response.Header.SetBytesKV(header[0], header[1])
	}

	// TODO only for debug mode
	context.Context.Response.Header.Set("X-Prudence-Cached", "true")

	// Body only in GET
	if context.Context.IsGet() {
		context.Context.Response.SetBody(self.Body)
	}
}

func (self *CacheEntry) Expired() bool {
	return time.Now().After(self.Expiration)
}

func copyBytes(bytes []byte) []byte {
	bytes_ := make([]byte, len(bytes))
	copy(bytes_, bytes)
	return bytes_
}
