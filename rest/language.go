package rest

import (
	"sort"
	"strings"
)

//
// Language
//

type Language struct {
	Name   string
	Tag    string
	SubTag string
}

func NewLanguage(name string) Language {
	self := Language{Name: name}
	self.Tag, self.SubTag, _ = strings.Cut(name, "-")
	return self
}

// ([fmt.Stringify] interface)
func (self Language) String() string {
	return self.Name
}

//
// LanguagePreference
//

type LanguagePreference struct {
	Language
	Weight float64
}

func ParseLanguagePreference(text string) (LanguagePreference, error) {
	if preference, err := ParsePreference(text); err == nil {
		return LanguagePreference{
			Language: NewLanguage(preference.Name),
			Weight:   preference.Weight,
		}, nil
	} else {
		return LanguagePreference{}, err
	}
}

func (self *LanguagePreference) Matches(language Language, anySubTag bool) bool {
	wildcard := self.Name == "*"

	if !wildcard {
		if !strings.EqualFold(self.Tag, language.Tag) {
			return false
		}
		if !anySubTag && !strings.EqualFold(self.SubTag, language.SubTag) {
			return false
		}
	}

	return true
}

// ([fmt.Stringify] interface)
func (self LanguagePreference) String() string {
	return Preference{self.Name, self.Weight}.String()
}

//
// LanguagePreferences
//

type LanguagePreferences []LanguagePreference

func ParseLanguagePreferences(text string) LanguagePreferences {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Language

	var self LanguagePreferences

	if text = strings.TrimSpace(text); text != "" {
		for _, text_ := range strings.Split(text, ",") {
			if languagePreference, err := ParseLanguagePreference(text_); err == nil {
				self = append(self, languagePreference)
			}
		}

		sort.Stable(sort.Reverse(self))
	}

	//log.Infof("%s", text)
	//log.Infof("%v", self)

	return self
}

// ([sort.Interface] interface)
func (self LanguagePreferences) Len() int {
	return len(self)
}

// ([sort.Interface] interface)
func (self LanguagePreferences) Less(i int, j int) bool {
	return self[i].Weight < self[j].Weight
}

// ([sort.Interface] interface)
func (self LanguagePreferences) Swap(i int, j int) {
	self[i], self[j] = self[j], self[i]
}
