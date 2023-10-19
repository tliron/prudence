package rest

import (
	"net/url"

	"github.com/tliron/commonlog"
)

var log = commonlog.GetLogger("prudence.rest")

const (
	HeaderAccept          = "Accept"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderAcceptLanguage  = "Accept-Language"
	HeaderCacheControl    = "Cache-Control"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentType     = "Content-Type"
	HeaderETag            = "ETag"
	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderIfNoneMatch     = "If-None-Match"
	HeaderLastModified    = "Last-Modified"
	HeaderLocation        = "Location"
	HeaderPrudenceCached  = "X-Prudence-Cached"
	HeaderServer          = "Server"
)

var DataContentTypes = []string{
	"application/yaml",
	"application/json",
	"application/xml",
	"application/cbor",
	"application/msgpack",
}

var EndRequest = struct{}{}

func CloneURLValues(values url.Values) url.Values {
	values_ := make(url.Values)
	for name, values__ := range values {
		values_[name] = append(values__[:0:0], values__...)
	}
	return values_
}
