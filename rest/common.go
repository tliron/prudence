package rest

import (
	"io"

	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("prudence.rest")

//
// Logger
//

type Logger struct{}

var logHttp = logging.GetLogger("prudence.http")

// fasthttp.Logger interface
func (self Logger) Printf(format string, args ...interface{}) {
	logHttp.Errorf(format, args...)
}

//
// WrappingWriter
//

type WrappingWriter interface {
	io.WriteCloser

	GetWrappedWriter() io.Writer
}
