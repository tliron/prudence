package platform

import (
	"errors"
	"fmt"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
)

type CreateFunc func(jsContext *commonjs.Context, config ard.StringMap) (any, error)

var Types = make(map[string]*Type)

//
// Type
//

type Type struct {
	Name string

	create    CreateFunc
	validKeys []string
}

func (self *Type) Create(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	if err := ValidateConfigKeys(self.Name, config, self.validKeys...); err != nil {
		return nil, err
	}

	return self.create(jsContext, config)
}

func RegisterType(name string, create CreateFunc, validKeys ...string) {
	Types[name] = &Type{
		Name:      name,
		create:    create,
		validKeys: validKeys,
	}
}

func GetType(name string) (*Type, error) {
	if type__, ok := Types[name]; ok {
		return type__, nil
	} else {
		return nil, fmt.Errorf("unsupported type: %s", name)
	}
}

func Create(jsContext *commonjs.Context, typeName string, config ard.StringMap) (any, error) {
	if type_, err := GetType(typeName); err == nil {
		return type_.Create(jsContext, config)
	} else {
		return nil, err
	}
}

func CreateFromConfig(jsContext *commonjs.Context, config ard.StringMap, defaultTypeName string) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	var typeName string
	var ok bool
	if typeName, ok = config_.Get("type").String(); ok {
		// Create a copy without "type" key
		config__ := make(ard.StringMap)
		for k, v := range config {
			if k != "type" {
				config__[k] = v
			}
		}
		config = config__
	} else if defaultTypeName != "" {
		typeName = defaultTypeName
	} else {
		return nil, errors.New("constructor config does not specify \"type\"")
	}

	if type_, err := GetType(typeName); err == nil {
		return type_.Create(jsContext, config)
	} else {
		return nil, err
	}
}

func CreateFromConfigList(jsContext *commonjs.Context, value ard.Value, typeName string, f func(instance any, config ard.StringMap)) error {
	if type_, err := GetType(typeName); err == nil {
		for _, config := range AsConfigList(value) {
			if config_, ok := ard.With(config).ConvertSimilar().NilMeansZero().StringMap(); ok {
				if instance, err := type_.Create(jsContext, config_); err == nil {
					f(instance, config_)
				} else {
					return err
				}
			} else {
				return fmt.Errorf("malformed constructor config: %T", config)
			}
		}
	} else {
		return err
	}

	return nil
}
