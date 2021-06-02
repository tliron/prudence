package platform

import (
	"errors"
	"fmt"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
)

type CreateFunc func(config ard.StringMap, context *js.Context) (interface{}, error)

var creators = make(map[string]CreateFunc)

func RegisterType(type_ string, create CreateFunc) {
	creators[type_] = create
}

func GetType(type_ string) (CreateFunc, error) {
	if create, ok := creators[type_]; ok {
		return create, nil
	} else {
		return nil, fmt.Errorf("unsupported \"type\": %s", type_)
	}
}

func OnTypes(f func(type_ string, create CreateFunc) bool) {
	for type_, create := range creators {
		if !f(type_, create) {
			return
		}
	}
}

func Create(config ard.StringMap, context *js.Context) (interface{}, error) {
	config_ := ard.NewNode(config)
	if type_, ok := config_.Get("type").String(false); ok {
		if create, err := GetType(type_); err == nil {
			return create(config, context)
		} else {
			return nil, err
		}
	} else {
		return nil, errors.New("\"type\" not specified")
	}
}
