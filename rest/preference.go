package rest

import (
	"strconv"
	"strings"
)

//
// Preference
//

type Preference struct {
	Name   string
	Weight float64
}

func ParsePreference(text string) (Preference, error) {
	self := Preference{Weight: 1.0}

	var annotation string
	var hasAnnotation bool
	if self.Name, annotation, hasAnnotation = strings.Cut(text, ";"); hasAnnotation {
		if strings.HasPrefix(annotation, "q=") {
			var err error
			if self.Weight, err = strconv.ParseFloat(annotation[2:], 64); err != nil {
				return self, err
			}
		}
	}

	return self, nil
}

// ([fmt.Stringify] interface)
func (self Preference) String() string {
	return self.Name + ";q=" + strconv.FormatFloat(self.Weight, 'f', -1, 64)
}
