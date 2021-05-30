package platform

import (
	"bytes"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

//
// CacheBackend
//

type CacheKey string

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

//
// CachedRepresentation
//

type CachedRepresentation struct {
	Groups     []CacheKey
	Headers    [][][]byte // list of key, value tuples
	Body       map[EncodingType][]byte
	Expiration time.Time
}

// fmt.Stringer interface
func (self *CachedRepresentation) String() string {
	keys := make([]string, 0, len(self.Body))
	for key := range self.Body {
		keys = append(keys, key.String())
	}
	return strings.Join(keys, ",")
}

func (self *CachedRepresentation) Expired() bool {
	return time.Now().After(self.Expiration)
}

// In seconds
func (self *CachedRepresentation) TimeToLive() float64 {
	duration := self.Expiration.Sub(time.Now()).Seconds()
	if duration < 0.0 {
		duration = 0.0
	}
	return duration
}

func (self *CachedRepresentation) GetBody(encoding EncodingType) ([]byte, bool) {
	if body, ok := self.Body[encoding]; ok {
		return body, false
	} else {
		switch encoding {
		case EncodingTypeBrotli:
			if plain, _ := self.GetBody(EncodingTypePlain); plain != nil {
				log.Debug("creating brotli body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteBrotli(buffer, plain)
				body = buffer.Bytes()
				self.Body[EncodingTypeBrotli] = body
				return body, true
			}

		case EncodingTypeGZip:
			if plain, _ := self.GetBody(EncodingTypePlain); plain != nil {
				log.Debug("creating gzip body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGzip(buffer, plain)
				body = buffer.Bytes()
				self.Body[EncodingTypeGZip] = body
				return body, true
			}

		case EncodingTypeDeflate:
			if plain, _ := self.GetBody(EncodingTypePlain); plain != nil {
				log.Debug("creating deflate body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteDeflate(buffer, plain)
				body = buffer.Bytes()
				self.Body[EncodingTypeDeflate] = body
				return body, true
			}

		case EncodingTypePlain:
			// Try decoding an existing body
			if deflate, ok := self.Body[EncodingTypeDeflate]; ok {
				log.Debug("creating plain body from default")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteInflate(buffer, deflate)
				body = buffer.Bytes()
				self.Body[EncodingTypePlain] = body
				return body, true
			} else if gzip, ok := self.Body[EncodingTypeGZip]; ok {
				log.Debug("creating plain body from gzip")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGunzip(buffer, gzip)
				body = buffer.Bytes()
				self.Body[EncodingTypePlain] = body
				return body, true
			} else if brotli, ok := self.Body[EncodingTypeBrotli]; ok {
				log.Debug("creating plain body from brotli")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteUnbrotli(buffer, brotli)
				body = buffer.Bytes()
				self.Body[EncodingTypePlain] = body
				return body, true
			}
		}

		return nil, false
	}
}

func (self *CachedRepresentation) Update(key CacheKey) {
	if cacheBackend := GetCacheBackend(); cacheBackend != nil {
		cacheBackend.StoreRepresentation(key, self)
		log.Debugf("cached representation updated: %s|%s", key, self)
	}
}
