package rest

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// TODO goroutine to prune cache

var cache sync.Map

func ToCache(context *Context) {
	cacheKey := NewCacheKey(context)
	cacheEntry := NewCacheEntry(context)
	cache.Store(cacheKey, cacheEntry)
	context.Log.Debugf("cache stored: %s", cacheKey)
}

func FromCache(context *Context) (*CacheEntry, bool) {
	cacheKey := NewCacheKey(context)
	context.Log.Debugf("trying cache: %s", cacheKey)
	if cacheEntry, ok := cache.Load(cacheKey); ok {
		cacheEntry_ := cacheEntry.(*CacheEntry)
		if cacheEntry_.Expired() {
			context.Log.Debugf("cache expired: %s", cacheKey)
			cache.Delete(cacheKey)
			return nil, false
		} else {
			context.Log.Debugf("cache hit: %s", cacheKey)
			return cacheEntry_, true
		}
	} else {
		context.Log.Debugf("not cached: %s", cacheKey)
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
	Key         string
	ContentType string
	CharSet     string
	Language    string
}

func NewCacheKey(context *Context) CacheKey {
	return CacheKey{
		Key:         context.CacheKey,
		ContentType: context.ContentType,
		CharSet:     context.CharSet,
		Language:    context.Language,
	}
}

// fmt.Stringer interface
func (self CacheKey) String() string {
	return fmt.Sprintf("%s, %s, %s, %s", self.Key, self.ContentType, self.CharSet, self.Language)
}

//
// CacheEntry
//

type CacheEntry struct {
	Headers    [][][]byte
	Body       map[string][]byte
	Expiration time.Time
}

func NewCacheEntry(context *Context) *CacheEntry {
	body := make(map[string][]byte)
	if context.context.Request.Header.IsGet() {
		// Body exists only in GET
		body[GetContentEncoding(context.context)] = copyBytes(context.context.Response.Body())
	}

	var headers [][][]byte
	context.context.Response.Header.VisitAll(func(key []byte, value []byte) {
		// This is an annoying way to get all headers, but unfortunately if we
		// get the entire header via Header() there is no API to set it correctly
		// in CacheEntry.Write
		switch string(key) {
		case fasthttp.HeaderServer, fasthttp.HeaderCacheControl:
			return
		}

		//context.Log.Debugf("header: %s", key)
		headers = append(headers, [][]byte{copyBytes(key), copyBytes(value)})
	})

	return &CacheEntry{
		Body:       body,
		Headers:    headers,
		Expiration: time.Now().Add(time.Duration(context.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func (self *CacheEntry) Write(context *Context) {
	context.context.Response.Reset()

	// Annoyingly these were re-enabled by Reset above
	context.context.Response.Header.DisableNormalizing()
	context.context.Response.Header.SetNoDefaultContentType(true)

	// Headers
	for _, header := range self.Headers {
		context.context.Response.Header.AddBytesKV(header[0], header[1])
	}

	eTag := GetETag(context.context)

	// New max-age
	maxAge := int(self.TimeToLive())
	AddCacheControl(context.context, fmt.Sprintf("max-age=%d", maxAge))

	/*else if eTag == "" {
		// Don't store and *also* invalidate the existing client cache
		AddCacheControl(context.context, "no-store,max-age=0")
	}*/

	// TODO only for debug mode
	context.context.Response.Header.Set("X-Prudence-Cached", context.CacheKey)

	// Conditional

	if IfNoneMatch(context.context, eTag) {
		// The following headers should have been set:
		// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
		context.context.NotModified()
		return
	}

	if !context.context.IfModifiedSince(GetLastModified(context.context)) {
		// The following headers should have been set:
		// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
		context.context.NotModified()
		return
	}

	// Body (not for HEAD)

	if !context.context.IsHead() {
		var body []byte

		if context.context.Request.Header.HasAcceptEncoding("gzip") {
			body = self.GetBody("gzip")
		} else {
			body = self.GetBody("")
		}

		context.context.Response.SetBody(body)
	}
}

func (self *CacheEntry) Expired() bool {
	return time.Now().After(self.Expiration)
}

// In seconds
func (self *CacheEntry) TimeToLive() float64 {
	duration := self.Expiration.Sub(time.Now()).Seconds()
	if duration < 0.0 {
		duration = 0.0
	}
	return duration
}

func (self *CacheEntry) GetBody(encoding string) []byte {
	var body []byte

	var ok bool
	switch encoding {
	case "gzip":
		if body, ok = self.Body["gzip"]; !ok {
			if plain, ok := self.Body[""]; ok {
				log.Debug("creating gzip body")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGzip(buffer, plain)
				body = buffer.Bytes()
				self.Body["gzip"] = body
			}
		}

	case "":
		if body, ok = self.Body[""]; !ok {
			if gzip, ok := self.Body["gzip"]; ok {
				log.Debug("creating plain body")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGunzip(buffer, gzip)
				body = buffer.Bytes()
				self.Body[""] = body
			}
		}
	}

	return body
}

func copyBytes(bytes []byte) []byte {
	bytes_ := make([]byte, len(bytes))
	copy(bytes_, bytes)
	return bytes_
}
