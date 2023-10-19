package platform

import (
	"fmt"

	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

func ValidateConfigKeys(name string, stringMap ard.StringMap, keys ...string) error {
	isValid := func(key string) bool {
		for _, key_ := range keys {
			if key_ == key {
				return true
			}
		}

		return false
	}

	var invalidKeys []string
	for key := range stringMap {
		if !isValid(key) {
			invalidKeys = append(invalidKeys, key)
		}
	}

	switch len(invalidKeys) {
	case 0:
	case 1:
		return fmt.Errorf("invalid config key for %s: %q", name, invalidKeys[0])
	default:
		return fmt.Errorf("invalid config keys for %s: %s", name, util.JoinQuote(invalidKeys, ", "))
	}

	return nil
}

func AsList(value ard.Value) ard.List {
	switch value_ := value.(type) {
	case ard.List:
		return value_
	default:
		return ard.List{value}
	}
}

func AsConfigList(value ard.Value) ard.List {
	switch value_ := value.(type) {
	case ard.List:
		return value_
	case ard.StringMap:
		return ard.List{value}
	default:
		return nil
	}
}

func AsStringList(node *ard.Node) []string {
	if list, ok := node.StringList(); ok {
		return list
	} else if node.Value != nil {
		string_, _ := node.String()
		return []string{string_}
	} else {
		return nil
	}
}
