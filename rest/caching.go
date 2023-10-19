package rest

import (
	"bytes"
	"strconv"
	"time"

	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

func (self *Context) NewCacheKey() platform.CacheKey {
	return platform.CacheKey(self.CacheKey + platform.CACHE_KEY_SEPARATOR + self.Response.ContentType + platform.CACHE_KEY_SEPARATOR + self.Response.CharSet + platform.CACHE_KEY_SEPARATOR + self.Response.Language)
}

func (self *Context) NewCachedRepresentation(withBody bool) *platform.CachedRepresentation {
	body := make(map[platform.EncodingType][]byte)

	if withBody {
		contentEncoding := self.Response.Header.Get(HeaderContentEncoding)
		if encoding := platform.GetEncodingFromHeader(contentEncoding); encoding != platform.EncodingTypeUnsupported {
			body[encoding] = self.Response.Buffer.Bytes()
		} else {
			self.Log.Warningf("unsupported encoding: %s", contentEncoding)
		}
	}

	headers := make(map[string][]string)
	for name, values := range self.Response.Header {
		switch name {
		case HeaderCacheControl, HeaderServer, HeaderPrudenceCached:
			// Skip
		default:
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
		Expiration: time.Now().Add(time.Duration(self.CacheDuration * float64(time.Second))),
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
		Expiration: time.Now().Add(time.Duration(self.CacheDuration * float64(time.Second))),
	}
}

func (self *Context) LoadCachedRepresentation() (platform.CacheKey, *platform.CachedRepresentation, bool) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		key := self.NewCacheKey()
		if cached, ok := cacheBackend.LoadRepresentation(key); ok {
			self.Log.Debug("hit",
				"_scope", "cache",
				"key", key,
				"encodings", cached.String(),
			)
			return key, cached, true
		} else {
			self.Log.Debug("miss",
				"_scope", "cache",
				"key", key,
			)
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
		self.Log.Debug("deleted",
			"_scope", "cache",
			"key", key,
		)
	}
}

func (self *Context) StoreCachedRepresentation(withBody bool) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		key := self.NewCacheKey()
		cached := self.NewCachedRepresentation(withBody)
		cacheBackend.StoreRepresentation(key, cached)
		self.Log.Debug("stored",
			"_scope", "cache",
			"key", key,
			"encodings", cached.String(),
		)
	}
}

func (self *Context) StoreCachedRepresentationFromBody(encoding platform.EncodingType, body []byte) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		key := self.NewCacheKey()
		cached := self.NewCachedRepresentationFromBody(encoding, body)
		cacheBackend.StoreRepresentation(key, cached)
		self.Log.Debug("stored",
			"_scope", "cache",
			"key", key,
			"encodings", cached.String(),
		)
	}
}

func (self *Context) GetCachedRepresentationBody(cached *platform.CachedRepresentation) ([]byte, platform.EncodingType, bool) {
	encodingPreferences := ParseEncodingPreferences(self.Request.Header.Get(HeaderAcceptEncoding))
	encoding := encodingPreferences.NegotiateBest(self)
	if body, changed := cached.GetBody(encoding); body != nil {
		return body, encoding, changed
	} else {
		return nil, platform.EncodingTypeUnsupported, false
	}
}

func (self *Context) PresentCachedRepresentation(cached *platform.CachedRepresentation, withBody bool) bool {
	self.Response.Reset()

	header := self.Response.Header
	if cached.Headers != nil {
		for name, values := range cached.Headers {
			for _, value := range values {
				header.Add(name, value)
			}
		}
	}

	if self.Debug {
		header.Set(HeaderPrudenceCached, self.CacheKey)
	}

	if self.isNotModified(true) {
		return false
	}

	// Match client-side caching with server-side caching
	maxAge := int64(cached.TimeToLive())
	self.Response.Header.Set(HeaderCacheControl, "max-age="+strconv.FormatInt(maxAge, 10))

	if withBody {
		body, encoding, changed := self.GetCachedRepresentationBody(cached)
		if encodingHeader := encoding.Header(); encodingHeader != "" {
			header.Set(HeaderContentEncoding, encodingHeader)
		}
		self.Response.Buffer = bytes.NewBuffer(body)
		return changed
	}

	return false
}

func (self *Context) WriteCachedRepresentation(cached *platform.CachedRepresentation) (bool, error) {
	if body, changed := cached.GetBody(platform.EncodingTypeIdentity); body != nil {
		return changed, self.Write(body)
	} else {
		return false, nil
	}
}

func (self *Context) UpdateCachedRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheBackend.StoreRepresentation(key, cached)
		self.Log.Debug("updated",
			"_scope", "cache",
			"key", key,
			"encodings", cached.String(),
		)
	}
}
