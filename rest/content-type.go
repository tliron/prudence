package rest

import (
	"sort"
	"strings"
)

//
// ContentType
//

type ContentType struct {
	Name    string
	Type    string
	SubType string
}

func NewContentType(name string) ContentType {
	self := ContentType{Name: name}
	self.Type, self.SubType, _ = strings.Cut(name, "/")
	return self
}

// ([fmt.Stringify] interface)
func (self ContentType) String() string {
	return self.Name
}

//
// ContentTypePreference
//

type ContentTypePreference struct {
	ContentType
	Weight float64
}

func ParseContentTypePreference(text string) (ContentTypePreference, error) {
	if preference, err := ParsePreference(text); err == nil {
		return ContentTypePreference{
			ContentType: NewContentType(preference.Name),
			Weight:      preference.Weight,
		}, nil
	} else {
		return ContentTypePreference{}, err
	}
}

func (self *ContentTypePreference) Matches(contentType ContentType) bool {
	typeWildcard := self.Type == "*"
	subTypeWildcard := self.SubType == "*"

	if !typeWildcard && !strings.EqualFold(self.Type, contentType.Type) {
		return false
	}

	if !subTypeWildcard && !strings.EqualFold(self.SubType, contentType.SubType) {
		return false
	}

	return true
}

// ([fmt.Stringify] interface)
func (self ContentTypePreference) String() string {
	return Preference{self.Name, self.Weight}.String()
}

//
// ContentTypePreferences
//

type ContentTypePreferences []ContentTypePreference

func ParseContentTypePreferences(text string) ContentTypePreferences {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept

	var self ContentTypePreferences

	if text = strings.TrimSpace(text); text != "" {
		for _, text_ := range strings.Split(text, ",") {
			if contentTypePreference, err := ParseContentTypePreference(text_); err == nil {
				self = append(self, contentTypePreference)
			}
		}

		sort.Stable(sort.Reverse(self))
	}

	//log.Infof("%s", text)
	//log.Infof("%v", self)

	return self
}

// ([sort.Interface] interface)
func (self ContentTypePreferences) Len() int {
	return len(self)
}

// ([sort.Interface] interface)
func (self ContentTypePreferences) Less(i int, j int) bool {
	return self[i].Weight < self[j].Weight
}

// ([sort.Interface] interface)
func (self ContentTypePreferences) Swap(i int, j int) {
	self[i], self[j] = self[j], self[i]
}
