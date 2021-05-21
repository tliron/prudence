package platform

import (
	"errors"
	"fmt"

	"github.com/tliron/kutil/ard"
)

type CreateFunc func(config ard.StringMap, getRelativeURL GetRelativeURL) (interface{}, error)

var creators = make(map[string]CreateFunc)

func RegisterType(type_ string, create CreateFunc) {
	creators[type_] = create
}

func GetCreator(type_ string) (CreateFunc, error) {
	if create, ok := creators[type_]; ok {
		return create, nil
	} else {
		return nil, fmt.Errorf("unsupported \"type\": %s", type_)
	}
}

func Create(config ard.StringMap, getRelativeURL GetRelativeURL) (interface{}, error) {
	config_ := ard.NewNode(config)
	if type_, ok := config_.Get("type").String(false); ok {
		if create, err := GetCreator(type_); err == nil {
			return create(config, getRelativeURL)
		} else {
			return nil, err
		}
	} else {
		return nil, errors.New("\"type\" not specified")
	}
}
