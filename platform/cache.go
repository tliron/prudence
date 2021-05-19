package platform

import (
	"fmt"
	"time"
)

//
// CacheBackend
//

var cacheBackend CacheBackend

type CacheBackend interface {
	Load(cacheKey CacheKey) (*CacheEntry, bool)      // sync
	Store(cacheKey CacheKey, cacheEntry *CacheEntry) // async
	Delete(cacheKey CacheKey)                        // async
}

func SetCacheBackend(cacheBackend_ CacheBackend) {
	cacheBackend = cacheBackend_
}

func GetCacheBackend() CacheBackend {
	return cacheBackend
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

// fmt.Stringer interface
func (self CacheKey) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", self.Key, self.ContentType, self.CharSet, self.Language)
}

//
// CacheEntry
//

type CacheEntry struct {
	Headers    [][][]byte              // list of key, value tuples
	Body       map[EncodingType][]byte // encoding type -> body
	Expiration time.Time
}

// fmt.Stringer interface
func (self *CacheEntry) String() string {
	keys := make([]string, 0, len(self.Body))
	for key := range self.Body {
		keys = append(keys, key.String())
	}
	return fmt.Sprintf("%s", keys)
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
