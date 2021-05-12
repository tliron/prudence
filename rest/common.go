package rest

import (
	"errors"
	"fmt"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/prudence/js/common"
)

var log = logging.GetLogger("prudence.rest")

var logHttp = logging.GetLogger("prudence.rest.http")

type Logger struct{}

// fasthttp.Logger interface
func (self Logger) Printf(format string, args ...interface{}) {
	logHttp.Errorf(format, args...)
}

type CreateFunc func(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error)

var createFuncs = make(map[string]CreateFunc)

func Register(type_ string, createFunc CreateFunc) {
	createFuncs[type_] = createFunc
}

func Create(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	config_ := ard.NewNode(config)
	if type_, ok := config_.Get("type").String(false); ok {
		if create, ok := createFuncs[type_]; ok {
			return create(config, getRelativeURL)
		} else {
			return nil, fmt.Errorf("unsupported \"type\": %s", type_)
		}
	} else {
		return nil, errors.New("\"type\" not specified")
	}
}
