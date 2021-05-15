package rest

import (
	"regexp"
	"strings"
)

const PATH_VARIABLE = "__path"

const PATH_VARIABLE_RE = `(?P<` + PATH_VARIABLE + `>.*)`

//
// PathTemplate
//
// Matches URI paths and extracts variables
//
// Variables are wrapped in "{" and "}" and do not extend beyond a path boundary ("/")
//
// The "*" wildcard extracts any character into the "PATH" variable
//

type PathTemplate struct {
	Template          string
	RegularExpression *regexp.Regexp
}

var PathTemplateAll = &PathTemplate{"", nil}

func NewPathTemplate(path string) *PathTemplate {
	// /resource/{name}
	// ^/resource/(?P<name>[^/]*)$

	if path == "" {
		// Empty template always matches
		return PathTemplateAll
	}

	var builder strings.Builder
	inVariable := false

	builder.WriteRune('^')

	for _, rune_ := range path {
		// TODO alow escaping

		if inVariable {
			switch rune_ {
			case '}':
				inVariable = false
				builder.WriteString(`>[^/]*)`)

			default:
				// Group name
				builder.WriteRune(rune_)
			}
		} else {
			switch rune_ {
			case '{':
				inVariable = true
				builder.WriteString(`(?P<`)

			case '*':
				builder.WriteString(PATH_VARIABLE_RE)

			default:
				builder.WriteString(regexp.QuoteMeta(string(rune_)))
			}
		}
	}

	builder.WriteRune('$')

	return &PathTemplate{
		Template:          path,
		RegularExpression: regexp.MustCompile(builder.String()),
	}
}

func (self *PathTemplate) Match(path string) map[string]string {
	if self.RegularExpression == nil {
		// Empty template always matches
		return make(map[string]string)
	}

	if matches := self.RegularExpression.FindStringSubmatch(path); matches != nil {
		//log.Debugf("matched: %s", self.RegularExpression.String())
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

//
// PathTemplates
//
// Matches any single template (in sequence)
//

type PathTemplates []*PathTemplate

func NewPathTemplates(paths []string) PathTemplates {
	self := make(PathTemplates, len(paths))
	for index, path := range paths {
		self[index] = NewPathTemplate(path)
	}
	return self
}

func (self PathTemplates) MatchAny(path string) map[string]string {
	for _, pathTemplate := range self {
		if matches := pathTemplate.Match(path); matches != nil {
			return matches
		}
	}

	return nil
}
