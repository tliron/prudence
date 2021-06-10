package rest

import (
	"fmt"
	"sort"
	"strconv"
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

	s := strings.SplitN(name, "/", 2)
	self.Type = s[0]
	if len(s) == 2 {
		self.SubType = s[1]
	}

	return self
}

// fmt.Stringify interface
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
	self := ContentTypePreference{Weight: 1.0}

	s := strings.SplitN(text, ";", 2)
	self.ContentType = NewContentType(s[0])

	// Annotation
	if len(s) == 2 {
		annotationText := s[1]
		if strings.HasPrefix(annotationText, "q=") {
			var err error
			if self.Weight, err = strconv.ParseFloat(annotationText[2:], 64); err != nil {
				return self, err
			}
		}
	}

	return self, nil
}

func (self *ContentTypePreference) Matches(contentType ContentType, matchWildcard bool) bool {
	typeWildcard := self.Type == "*"
	subTypeWildcard := self.SubType == "*"

	if !matchWildcard && typeWildcard && subTypeWildcard {
		return false
	}

	if !typeWildcard && (self.Type != contentType.Type) {
		return false
	}

	if !subTypeWildcard && (self.SubType != contentType.SubType) {
		return false
	}

	return true
}

// fmt.Stringify interface
func (self ContentTypePreference) String() string {
	return fmt.Sprintf("%s;q=%g", self.Name, self.Weight)
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

// sort.Interface interface
func (self ContentTypePreferences) Len() int {
	return len(self)
}

// sort.Interface interface
func (self ContentTypePreferences) Less(i int, j int) bool {
	return self[i].Weight < self[j].Weight
}

// sort.Interface interface
func (self ContentTypePreferences) Swap(i int, j int) {
	self[i], self[j] = self[j], self[i]
}
