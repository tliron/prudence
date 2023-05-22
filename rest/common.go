package rest

import (
	"io"

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

//
// WrappingWriter
//

type WrappingWriter interface {
	io.WriteCloser

	GetWrappedWriter() io.Writer
}

// Utils

func copyBytes(bytes []byte) []byte {
	bytes_ := make([]byte, len(bytes))
	copy(bytes_, bytes)
	return bytes_
}
