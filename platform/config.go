package platform

import (
	"github.com/tliron/kutil/ard"
)

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

func AsStringList(value ard.Value) []string {
	switch value_ := value.(type) {
	case ard.List:
		return ToStringList(value_)
	case string:
		return []string{value_}
	default:
		return nil
	}
}

func ToStringList(list ard.List) []string {
	stringList := make([]string, 0, len(list))
	for _, element := range list {
		if element_, ok := element.(string); ok {
			stringList = append(stringList, element_)
		}
	}
	return stringList
}
