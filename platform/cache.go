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
		// We don't have our encoding, so convert an existing one
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

		case EncodingTypeDeflate:
			if plain, _ := self.GetBody(EncodingTypeIdentity); plain != nil {
				log.Debug("creating deflate body from plain")
				buffer := bytes.NewBuffer(nil)
				if err := EncodeDeflate(plain, buffer); err == nil {
					body = buffer.Bytes()
					self.Body[EncodingTypeDeflate] = body
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

		case EncodingTypeIdentity:
			// Try decoding an existing body
			// TODO: we should try these in descending order of decoding performance
			if brotli, ok := self.Body[EncodingTypeBrotli]; ok {
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
			} else if deflate, ok := self.Body[EncodingTypeDeflate]; ok {
				log.Debug("creating plain body from deflate")
				buffer := bytes.NewBuffer(nil)
				if err := DecodeDeflate(deflate, buffer); err == nil {
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
			}
		}

		return nil, false
	}
}

func (self *CachedRepresentation) GetHeadersSize() int {
	var size int
	for _, headers := range self.Headers {
		for _, header := range headers {
			size += len(header)
		}
	}
	return size
}

func (self *CachedRepresentation) GetBodySize() int {
	var size int
	for _, body := range self.Body {
		size += len(body)
	}
	return size
}

func (self *CachedRepresentation) GetSize() int {
	var size int
	for _, group := range self.Groups {
		size += len(group)
	}
	size += self.GetHeadersSize()
	size += self.GetBodySize()
	return size
}

func (self *CachedRepresentation) Update(key CacheKey) {
	if cacheBackend := GetCacheBackend(); cacheBackend != nil {
		cacheBackend.StoreRepresentation(key, self)
		log.Debugf("cached representation updated: %s|%s", key, self)
	}
}
