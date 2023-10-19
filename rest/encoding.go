package rest

import (
	"net/http"
	"sort"
	"strings"

	"github.com/tliron/prudence/platform"
)

func SetBestEncodeWriter(restContext *Context) bool {
	encodingPreferences := ParseEncodingPreferences(restContext.Request.Header.Get(HeaderAcceptEncoding))
	encoding := encodingPreferences.NegotiateBest(restContext)

	switch encoding {
	case platform.EncodingTypeIdentity:
		return true

	case platform.EncodingTypeBrotli, platform.EncodingTypeDeflate, platform.EncodingTypeGZip, platform.EncodingTypeZstandard:
		restContext.Response.Header.Set(HeaderContentEncoding, encoding.Header())

	default:
		restContext.Response.Status = http.StatusNotAcceptable
		return false
	}

	restContext.Writer = encoding.NewWriter(restContext.Writer)
	return true
}

//
// EncodingPreference
//

type EncodingPreference struct {
	Preference
	Type platform.EncodingType
}

func ParseEncodingPreference(text string) (EncodingPreference, error) {
	var self EncodingPreference
	var err error
	if self.Preference, err = ParsePreference(text); err == nil {
		self.Type = platform.GetEncodingFromHeader(self.Name)
		return self, nil
	} else {
		return self, err
	}
}

//
// EncodingPreferences
//

type EncodingPreferences []EncodingPreference

func ParseEncodingPreferences(text string) EncodingPreferences {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Encoding

	var self EncodingPreferences

	if text = strings.TrimSpace(text); text != "" {
		for _, text_ := range strings.Split(text, ",") {
			text_ = strings.TrimSpace(text_)
			if encodingPreference, err := ParseEncodingPreference(text_); err == nil {
				self = append(self, encodingPreference)
			}
		}

		sort.Stable(sort.Reverse(self))
	}

	//log.Infof("%s", text)
	//log.Infof("%v", self)

	return self
}

func (self EncodingPreferences) ForbidIdentity() bool {
	for _, encodingPreference := range self {
		if (encodingPreference.Type == platform.EncodingTypeIdentity) || (encodingPreference.Name == "*") {
			if encodingPreference.Weight == 0.0 {
				return true
			}
		}
	}

	return false
}

func (self EncodingPreferences) NegotiateBest(restContext *Context) platform.EncodingType {
	for _, encodingPreference := range self {
		if encodingPreference.Weight != 0.0 {
			switch encodingPreference.Type {
			// Note: "compress" has been deprecated
			case platform.EncodingTypeUnsupported, platform.EncodingTypeCompress:
			default:
				return encodingPreference.Type
			}
		}
	}

	if !self.ForbidIdentity() {
		return platform.EncodingTypeIdentity
	} else {
		return platform.EncodingTypeUnsupported
	}
}

// ([sort.Interface] interface)
func (self EncodingPreferences) Len() int {
	return len(self)
}

// ([sort.Interface] interface)
func (self EncodingPreferences) Less(i int, j int) bool {
	return self[i].Weight < self[j].Weight
}

// ([sort.Interface] interface)
func (self EncodingPreferences) Swap(i int, j int) {
	self[i], self[j] = self[j], self[i]
}
