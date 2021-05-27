package rest

import (
	"fmt"
	"time"

	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

func (self *Context) NewCacheKey() platform.CacheKey {
	return platform.CacheKey{
		Key:         self.CacheKey,
		ContentType: self.ContentType,
		CharSet:     self.CharSet,
		Language:    self.Language,
	}
}

func (self *Context) NewCachedRepresentation(withBody bool) *platform.CachedRepresentation {
	body := make(map[platform.EncodingType][]byte)
	if withBody {
		// Body exists only in GET
		contentEncoding := GetContentEncoding(self.Context)
		if encodingType := GetEncodingType(contentEncoding); encodingType != platform.EncodingTypeUnsupported {
			body[encodingType] = copyBytes(self.Context.Response.Body())
		} else {
			self.Log.Warningf("unsupported encoding: %s", contentEncoding)
		}
	}

	// This is an annoying way to get all headers, but unfortunately if we
	// get the entire header via Header() there is no API to set it correctly
	// in CacheEntry.Write
	var headers [][][]byte
	self.Context.Response.Header.VisitAll(func(key []byte, value []byte) {
		switch string(key) {
		case fasthttp.HeaderServer, fasthttp.HeaderCacheControl:
			return
		}

		//context.Log.Debugf("header: %s", key)
		headers = append(headers, [][]byte{copyBytes(key), copyBytes(value)})
	})

	groups := make([]string, len(self.CacheGroups))
	copy(groups, self.CacheGroups)

	return &platform.CachedRepresentation{
		Groups:     groups,
		Body:       body,
		Headers:    headers,
		Expiration: time.Now().Add(time.Duration(self.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func (self *Context) NewCachedRepresentationFromBody(encoding platform.EncodingType, body []byte) *platform.CachedRepresentation {
	groups := make([]string, len(self.CacheGroups))
	copy(groups, self.CacheGroups)

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
			self.Log.Debugf("cache hit: %s|%s", key, cached)
			return key, cached, true
		} else {
			self.Log.Debugf("cache miss: %s", key)
			return platform.CacheKey{}, nil, false
		}
	} else {
		return platform.CacheKey{}, nil, false
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
	if self.Context.Request.Header.HasAcceptEncoding("br") {
		return cached.GetBody(platform.EncodingTypeBrotli)
	} else if self.Context.Request.Header.HasAcceptEncoding("gzip") {
		return cached.GetBody(platform.EncodingTypeGZip)
	} else if self.Context.Request.Header.HasAcceptEncoding("deflate") {
		return cached.GetBody(platform.EncodingTypeDeflate)
	} else {
		return cached.GetBody(platform.EncodingTypePlain)
	}
}

func (self *Context) DescribeCachedRepresentation(cached *platform.CachedRepresentation) bool {
	self.Context.Response.Reset()

	// Annoyingly these were re-enabled by Reset above
	self.Context.Response.Header.DisableNormalizing()
	self.Context.Response.Header.SetNoDefaultContentType(true)

	if self.Debug {
		self.Context.Response.Header.Set(CACHED_HEADER, self.CacheKey)
	}

	for _, header := range cached.Headers {
		self.Context.Response.Header.AddBytesKV(header[0], header[1])
	}

	return !self.isNotModified()
}

func (self *Context) PresentCachedRepresentation(cached *platform.CachedRepresentation, withBody bool) bool {
	if !self.DescribeCachedRepresentation(cached) {
		return false
	}

	// Match client-side caching with server-side caching
	maxAge := int(cached.TimeToLive())
	AddCacheControl(self.Context, fmt.Sprintf("max-age=%d", maxAge))

	if withBody {
		body, changed := self.GetCachedRepresentationBody(cached)
		self.Context.Response.SetBody(body)
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
