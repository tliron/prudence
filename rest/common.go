package rest

import (
	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("prudence.rest")

var logHttp = logging.GetLogger("prudence.rest.http")

type Logger struct{}

// fasthttp.Logger interface
func (self Logger) Printf(format string, args ...interface{}) {
	logHttp.Errorf(format, args...)
}
