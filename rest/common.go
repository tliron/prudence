package rest

import (
	"io"

	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("prudence.rest")

var logCache = logging.GetLogger("prudence.cache")

const (
	CACHED_HEADER = "X-Prudence-Cached"
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
