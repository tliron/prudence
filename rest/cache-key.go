package rest

import (
	"github.com/tliron/prudence/platform"
)

func NewCacheKey(context *Context) platform.CacheKey {
	return platform.CacheKey{
		Key:         context.CacheKey,
		ContentType: context.ContentType,
		CharSet:     context.CharSet,
		Language:    context.Language,
	}
}
