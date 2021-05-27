package rest

import (
	"io"

	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("prudence.rest")

var logCache = logging.GetLogger("prudence.cache")

const (
	PATH_HEADER   = "X-Prudence-Path"
	CACHED_HEADER = "X-Prudence-Cached"
)

//
// Logger
//

type Logger struct {
	log logging.Logger
}

// fasthttp.Logger interface
func (self Logger) Printf(format string, args ...interface{}) {
	self.log.Errorf(format, args...)
}

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
