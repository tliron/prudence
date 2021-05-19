package js

import (
	"fmt"
	"io"
	"strings"

	"github.com/beevik/etree"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/util"
	"github.com/tliron/yamlkeys"
)

func (self *PrudenceAPI) ValidateFormat(code string, format string) error {
	return formatpkg.Validate(code, format)
}

func (self *PrudenceAPI) Decode(code string, format string, all bool) (ard.Value, error) {
	switch format {
	case "yaml", "":
		if all {
			if value, err := yamlkeys.DecodeAll(strings.NewReader(code)); err == nil {
				value_, _ := ard.MapsToStringMaps(value)
				return value_, err
			} else {
				return nil, err
			}
		} else {
			value, _, err := ard.DecodeYAML(code, false)
			value, _ = ard.MapsToStringMaps(value)
			return value, err
		}

	case "json":
		value, _, err := ard.DecodeJSON(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	case "cjson":
		value, _, err := ard.DecodeCompatibleJSON(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	case "xml":
		value, _, err := ard.DecodeCompatibleXML(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	case "cbor":
		value, _, err := ard.DecodeCBOR(code, false)
		value, _ = ard.MapsToStringMaps(value)
		return value, err

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (self *PrudenceAPI) Encode(value interface{}, format string, indent string, writer io.Writer) (string, error) {
	if writer == nil {
		return formatpkg.Encode(value, format, false)
	} else {
		err := formatpkg.Write(value, format, indent, false, writer)
		return "", err
	}
}

func (self *PrudenceAPI) NewXMLDocument() *etree.Document {
	return etree.NewDocument()
}

// Encode bytes as base64
func (self *PrudenceAPI) Btoa(bytes []byte) string {
	return util.ToBase64(bytes)
}

// Decode base64 to bytes
func (self *PrudenceAPI) Atob(b64 string) ([]byte, error) {
	// Note: if you need a string in JavaScript: String.fromCharCode.apply(null, .atob(...))
	return util.FromBase64(b64)
}
