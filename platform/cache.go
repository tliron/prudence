package platform

import (
	"bytes"
	"strings"
	"time"
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
	Headers    map[string][]string
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
			if plain, _ := self.GetBody(EncodingTypeIdentity); plain != nil {
				log.Debug("creating brotli body from plain")
				buffer := bytes.NewBuffer(nil)
				if err := EncodeBrotli(plain, buffer); err == nil {
					body = buffer.Bytes()
					self.Body[EncodingTypeBrotli] = body
					return body, true
				} else {
					log.Errorf("%s", err)
					return nil, false
				}
			}

		case EncodingTypeGZip:
			if plain, _ := self.GetBody(EncodingTypeIdentity); plain != nil {
				log.Debug("creating gzip body from plain")
				buffer := bytes.NewBuffer(nil)
				if err := EncodeGZip(plain, buffer); err == nil {
					body = buffer.Bytes()
					self.Body[EncodingTypeGZip] = body
					return body, true
				} else {
					log.Errorf("%s", err)
					return nil, false
				}
			}

		case EncodingTypeFlate:
			if plain, _ := self.GetBody(EncodingTypeIdentity); plain != nil {
				log.Debug("creating flate body from plain")
				buffer := bytes.NewBuffer(nil)
				if err := EncodeFlate(plain, buffer); err == nil {
					body = buffer.Bytes()
					self.Body[EncodingTypeFlate] = body
					return body, true
				} else {
					log.Errorf("%s", err)
					return nil, false
				}
			}

		case EncodingTypeIdentity:
			// Try decoding an existing body
			if deflate, ok := self.Body[EncodingTypeFlate]; ok {
				log.Debug("creating plain body from flate")
				buffer := bytes.NewBuffer(nil)
				if err := DecodeFlate(deflate, buffer); err == nil {
					body = buffer.Bytes()
					self.Body[EncodingTypeIdentity] = body
					return body, true
				} else {
					log.Errorf("%s", err)
					return nil, false
				}
			} else if gzip, ok := self.Body[EncodingTypeGZip]; ok {
				log.Debug("creating plain body from gzip")
				buffer := bytes.NewBuffer(nil)
				if err := DecodeGZip(gzip, buffer); err == nil {
					body = buffer.Bytes()
					self.Body[EncodingTypeIdentity] = body
					return body, true
				} else {
					log.Errorf("%s", err)
					return nil, false
				}
			} else if brotli, ok := self.Body[EncodingTypeBrotli]; ok {
				log.Debug("creating plain body from brotli")
				buffer := bytes.NewBuffer(nil)
				if err := DecodeBrotli(brotli, buffer); err == nil {
					body = buffer.Bytes()
					self.Body[EncodingTypeIdentity] = body
					return body, true
				} else {
					log.Errorf("%s", err)
					return nil, false
				}
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
