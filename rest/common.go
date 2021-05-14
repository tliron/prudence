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

type Startable interface {
	Start() error
	Stop() error
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

func asConfigList(value ard.Value) ard.List {
	switch value_ := value.(type) {
	case ard.List:
		return value_
	case ard.StringMap:
		return ard.List{value}
	default:
		return nil
	}
}

func asStringList(value ard.Value) []string {
	switch value_ := value.(type) {
	case ard.List:
		return toStringList(value_)
	case string:
		return []string{value_}
	default:
		return nil
	}
}

func toStringList(list ard.List) []string {
	stringList := make([]string, 0, len(list))
	for _, element := range list {
		if element_, ok := element.(string); ok {
			stringList = append(stringList, element_)
		}
	}
	return stringList
}
