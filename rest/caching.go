package rest

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/platform"
)

func (self *Context) NewCacheKey() platform.CacheKey {
	return platform.CacheKey(fmt.Sprintf("%s|%s|%s|%s", self.CacheKey, self.Response.ContentType, self.Response.CharSet, self.Response.Language))
}

func (self *Context) NewCachedRepresentation(withBody bool) *platform.CachedRepresentation {
	body := make(map[platform.EncodingType][]byte)

	if withBody {
		contentEncoding := self.Response.Header.Get("Content-Encoding")
		if encodingType := GetEncodingType(contentEncoding); encodingType != platform.EncodingTypeUnsupported {
			body[encodingType] = copyBytes(self.Response.Buffer.Bytes())
		} else {
			self.Log.Warningf("unsupported encoding: %s", contentEncoding)
		}
	}

	headers := make(map[string][]string)
	for name, values := range self.Response.Header {
		if name != "Cache-Control" {
			headers[name] = ard.Copy(values).([]string)
		}
	}

	groups := make([]platform.CacheKey, len(self.CacheGroups))
	for index, group := range self.CacheGroups {
		groups[index] = platform.CacheKey(group)
	}

	return &platform.CachedRepresentation{
		Groups:     groups,
		Body:       body,
		Headers:    headers,
		Expiration: time.Now().Add(time.Duration(self.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func (self *Context) NewCachedRepresentationFromBody(encoding platform.EncodingType, body []byte) *platform.CachedRepresentation {
	groups := make([]platform.CacheKey, len(self.CacheGroups))
	for index, group := range self.CacheGroups {
		groups[index] = platform.CacheKey(group)
	}

	return &platform.CachedRepresentation{
		Groups:     groups,
		Body:       map[platform.EncodingType][]byte{encoding: body},
		Headers:    nil,
		Expiration: time.Now().Add(time.Duration(self.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func (self *Context) LoadCachedRepresentation() (platform.CacheKey, *platform.CachedRepresentation, bool) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		key := self.NewCacheKey()
		if cached, ok := cacheBackend.LoadRepresentation(key); ok {
			self.Log.Debugf("cache hit: %s, %s", key, cached)
			return key, cached, true
		} else {
			self.Log.Debugf("cache miss: %s", key)
			return "", nil, false
		}
	} else {
		return "", nil, false
	}
}

func (self *Context) DeleteCachedRepresentation() {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		key := self.NewCacheKey()
		cacheBackend.DeleteRepresentation(key)
		logCache.Debugf("representation deleted: %s", key)
	}
}

func (self *Context) StoreCachedRepresentation(withBody bool) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		key := self.NewCacheKey()
		cached := self.NewCachedRepresentation(withBody)
		cacheBackend.StoreRepresentation(key, cached)
		self.Log.Debugf("representation stored: %s|%s", key, cached)
	}
}

func (self *Context) StoreCachedRepresentationFromBody(encoding platform.EncodingType, body []byte) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		key := self.NewCacheKey()
		cached := self.NewCachedRepresentationFromBody(encoding, body)
		cacheBackend.StoreRepresentation(key, cached)
		self.Log.Debugf("representation stored: %s|%s", key, cached)
	}
}

func (self *Context) GetCachedRepresentationBody(cached *platform.CachedRepresentation) ([]byte, bool) {
	return cached.GetBody(NegotiateBestEncodingType(self.Request.Header))
}

func (self *Context) DescribeCachedRepresentation(cached *platform.CachedRepresentation) bool {
	self.Response.Buffer.Reset()

	if self.Debug {
		self.Response.Header.Set(CACHED_HEADER, self.CacheKey)
	}

	if cached.Headers != nil {
		for name, values := range cached.Headers {
			for _, value := range values {
				self.Response.Header.Add(name, value)
			}
		}
	}

	return !self.isNotModified()
}

func (self *Context) PresentCachedRepresentation(cached *platform.CachedRepresentation, withBody bool) bool {
	if !self.DescribeCachedRepresentation(cached) {
		return false
	}

	// Match client-side caching with server-side caching
	maxAge := int(cached.TimeToLive())
	self.Response.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))

	if withBody {
		body, changed := self.GetCachedRepresentationBody(cached)
		self.Response.Buffer = bytes.NewBuffer(body)
		return changed
	}

	return false
}

func (self *Context) WriteCachedRepresentation(cached *platform.CachedRepresentation) (bool, int, error) {
	if body, changed := cached.GetBody(platform.EncodingTypePlain); body != nil {
		n, err := self.Write(body)
		return changed, n, err
	} else {
		return false, 0, nil
	}
}
