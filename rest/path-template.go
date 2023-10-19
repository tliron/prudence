package rest

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	PathVariable   = "__path"
	PathVariableRe = `(?P<` + PathVariable + `>.*)`
)

//
// PathTemplate
//
// Matches URI paths and extracts variables.
//
// Variables are wrapped in "{" and "}" and do not extend beyond a slash, unless "*"
// appears before the "}".
//
// The "*" wildcard matches any characters into the "__path" variable. It can only be
// used once per path.
//

type PathTemplate struct {
	Template                               string
	RegularExpression                      *regexp.Regexp
	RedirectTrailingSlashRegularExpression *regexp.Regexp
}

var PathTemplateAll = &PathTemplate{"", nil, nil}

func NewPathTemplate(path string) (*PathTemplate, error) {
	if path == "" {
		// Empty template always matches
		return PathTemplateAll, nil
	}

	var builder strings.Builder
	var inVariable bool
	var containsWildcard bool
	var variableContainsWildcard bool
	var redirectTrailingSlash string
	var skipNext bool

	builder.WriteRune('^')

	runes := []rune(path)
	last := len(runes) - 1
	for index, rune_ := range runes {
		// TODO alow escaping

		if skipNext {
			skipNext = false
			continue
		}

		if inVariable {
			switch rune_ {
			case '}':
				if variableContainsWildcard {
					builder.WriteString(`>.*)`)
					variableContainsWildcard = false
				} else {
					builder.WriteString(`>[^/]*)`)
				}
				inVariable = false

			case '*':
				variableContainsWildcard = true

			default:
				if variableContainsWildcard {
					return nil, fmt.Errorf("variable name in path template has \"*\" but not as last character: %s", path)
				}

				// Group name
				builder.WriteRune(rune_)
			}
		} else {
			switch rune_ {
			case '{':
				inVariable = true
				builder.WriteString(`(?P<`)

			case '}':
				return nil, fmt.Errorf("path template contains \"}\" without a \"{\" before it: %s", path)

			case '*':
				if containsWildcard {
					return nil, fmt.Errorf("path template contains more than one \"*\": %s", path)
				}
				builder.WriteString(PathVariableRe)
				containsWildcard = true

			case '/':
				if (index < last) && (runes[index+1] == '/') {
					if redirectTrailingSlash != "" {
						return nil, fmt.Errorf("path template contains more than \"//\": %s", path)
					}
					redirectTrailingSlash = builder.String() + "$"
					skipNext = true
				}
				builder.WriteRune('/')

			default:
				builder.WriteString(regexp.QuoteMeta(string(rune_)))
			}
		}
	}

	builder.WriteRune('$')

	if re, err := regexp.Compile(builder.String()); err == nil {
		var redirectTrailingSlashRe *regexp.Regexp
		if redirectTrailingSlash != "" {
			if redirectTrailingSlashRe, err = regexp.Compile(redirectTrailingSlash); err != nil {
				return nil, err
			}
		}

		return &PathTemplate{
			Template:                               path,
			RegularExpression:                      re,
			RedirectTrailingSlashRegularExpression: redirectTrailingSlashRe,
		}, nil
	} else {
		return nil, err
	}
}

func (self *PathTemplate) Match(path string) map[string]string {
	if self.RegularExpression == nil {
		// Empty template always matches
		return make(map[string]string)
	}

	if matches := self.RegularExpression.FindStringSubmatch(path); matches != nil {
		names := self.RegularExpression.SubexpNames()
		map_ := make(map[string]string)
		for index, match := range matches {
			if index > 0 {
				map_[names[index]] = match
			}
		}
		return map_
	}

	return nil
}

func (self *PathTemplate) MatchRedirectTrailingSlash(path string) bool {
	if self.RedirectTrailingSlashRegularExpression != nil {
		return self.RedirectTrailingSlashRegularExpression.MatchString(path)
	} else {
		return false
	}
}

// ([fmt.Stringify] interface)
func (self *PathTemplate) String() string {
	if self.RegularExpression != nil {
		return self.RegularExpression.String()
	} else {
		return ""
	}
}

//
// PathTemplates
//
// Matches any single template (in sequence)
//

type PathTemplates []*PathTemplate

func NewPathTemplates(paths ...string) (PathTemplates, error) {
	self := make(PathTemplates, len(paths))
	for index, path := range paths {
		if pathTemplate, err := NewPathTemplate(path); err == nil {
			self[index] = pathTemplate
		} else {
			return nil, err
		}
	}
	return self, nil
}

func (self PathTemplates) MatchAny(path string) map[string]string {
	for _, pathTemplate := range self {
		if matches := pathTemplate.Match(path); matches != nil {
			return matches
		}
	}

	return nil
}

func (self PathTemplates) MatchAnyRedirectTrailingSlash(path string) bool {
	for _, pathTemplate := range self {
		if pathTemplate.MatchRedirectTrailingSlash(path) {
			return true
		}
	}

	return false
}

// ([fmt.Stringify] interface)
func (self PathTemplates) String() string {
	var builder strings.Builder
	var last = len(self) - 1
	for index, pathTemplate := range self {
		builder.WriteString(pathTemplate.String())
		if index != last {
			builder.WriteString(", ")
		}
	}
	return builder.String()
}
