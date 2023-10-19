package platform

import (
	"strings"
	"time"
)

//
// CachedRepresentation
//

type CachedRepresentation struct {
	Groups     []CacheKey
	Headers    map[string][]string
	Body       map[EncodingType][]byte
	Expiration time.Time
}

// ([fmt.Stringer] interface)
func (self *CachedRepresentation) String() string {
	keys := make([]string, len(self.Body))
	index := 0
	for key := range self.Body {
		keys[index] = key.String()
		index++
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
	}

	// Try to reencode from other encodings (in order of decoding performance)
	if body, ok := self.ReencodeBody(EncodingTypeIdentity, encoding); ok {
		return body, true
	}
	if body, ok := self.ReencodeBody(EncodingTypeZstandard, encoding); ok {
		return body, true
	}
	if body, ok := self.ReencodeBody(EncodingTypeGZip, encoding); ok {
		return body, true
	}
	if body, ok := self.ReencodeBody(EncodingTypeDeflate, encoding); ok {
		return body, true
	}
	if body, ok := self.ReencodeBody(EncodingTypeBrotli, encoding); ok {
		return body, true
	}

	return nil, false
}

func (self *CachedRepresentation) ReencodeBody(fromEncoding EncodingType, toEncoding EncodingType) ([]byte, bool) {
	if fromEncoding != toEncoding {
		if decodedBody, ok := self.DecodeBody(fromEncoding); ok {
			if reencodedBody, err := toEncoding.Encoded(decodedBody); err == nil {
				self.Body[toEncoding] = reencodedBody
				return reencodedBody, true
			} else {
				log.Error(err.Error(), "_scope", "cache")
			}
		}
	}

	return nil, false
}

func (self *CachedRepresentation) DecodeBody(encoding EncodingType) ([]byte, bool) {
	if body, ok := self.Body[encoding]; ok {
		if decodedBody, err := encoding.Decoded(body); err == nil {
			return decodedBody, true
		} else {
			log.Error(err.Error(), "_scope", "cache")
		}
	}

	return nil, false
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
