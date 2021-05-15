package rest

import "fmt"

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
