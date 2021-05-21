package js

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/util"
)

func (self *PrudenceAPI) DeepCopy(value ard.Value) ard.Value {
	return ard.Copy(value)
}

func (self *PrudenceAPI) DeepEquals(a ard.Value, b ard.Value) bool {
	return ard.Equals(a, b)
}

func (self *PrudenceAPI) StringToBytes(string_ string) []byte {
	return util.StringToBytes(string_)
}

func (self *PrudenceAPI) BytesToString(bytes []byte) string {
	return util.BytesToString(bytes)
}

func (self *PrudenceAPI) Hash(value ard.Value) (string, error) {
	if hash, err := hashstructure.Hash(value, hashstructure.FormatV2, nil); err == nil {
		return strconv.FormatUint(hash, 10), nil
	} else {
		return "", err
	}
}

func (self *PrudenceAPI) Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (self *PrudenceAPI) JoinFilePath(elements ...string) string {
	return filepath.Join(elements...)
}

func (self *PrudenceAPI) IsType(value ard.Value, type_ string) (bool, error) {
	// Special case whereby an integer stored as a float type has been optimized to an integer type
	if (type_ == "!!float") && ard.IsInteger(value) {
		return true, nil
	}

	if validate, ok := ard.TypeValidators[ard.TypeName(type_)]; ok {
		return validate(value), nil
	} else {
		return false, fmt.Errorf("unsupported type: %s", type_)
	}
}

func (self *PrudenceAPI) Timestamp() ard.Value {
	return util.Timestamp(false)
}
