package jst

import (
	"fmt"
	"strings"

	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterRenderer("jst", RenderJST)
}

// platform.RenderFunc signature
func RenderJST(content string, jsContext *js.Context) (string, error) {
	if tags, final, err := getTags(content); err == nil {
		var jstContext platform.JSTContext

		jstContext.Builder.WriteString("exports.present = function(context) {\n")

		if len(tags) == 0 {
			// Optimize for when there are no tags
			jstContext.WriteLiteral(content)
		} else {
			last := 0

			for _, tag := range tags {
				// Previous chunk
				jstContext.WriteLiteral(content[last:tag.start])
				last = tag.end

				code := content[tag.start+2 : tag.end-2]
				trimmedCode := strings.TrimSpace(code)

				if trimmedCode == "" {
					continue
				}

				// Swallow trailing newline by default
				swallowTrailingNewline := true

				if content[tag.end-3] == '/' {
					// Disable the swallowing of trailing newline
					code = code[:len(code)-1]
					swallowTrailingNewline = false
				}

				// Handle
				handled := false
				platform.OnTags(func(prefix string, handleTag platform.HandleTagFunc) bool {
					if strings.HasPrefix(code, prefix) {
						if handleTag(&jstContext, code) {
							swallowTrailingNewline = false
						}
						handled = true
						return false
					}
					return true
				})

				if !handled {
					// Scriptlet tag
					jstContext.Builder.WriteString(trimmedCode)
					jstContext.Builder.WriteRune('\n')
				}

				if swallowTrailingNewline {
					// Skip trailing newline
					if (tag.end <= final) && (content[tag.end] == '\n') {
						last++
					}
				}
			}

			if last <= final {
				// Leftover chunk
				jstContext.WriteLiteral(content[last:])
			}
		}

		jstContext.Builder.WriteString("};\n")

		string_ := jstContext.Builder.String()
		//log.Debugf("%s", string_)

		return string_, nil
	} else {
		return "", err
	}
}

type tag struct {
	start int
	end   int
}

func getTags(content string) ([]tag, int, error) {
	var tags []tag
	final := len(content) - 1
	start := -1

	for index, rune_ := range content {
		switch rune_ {
		case '<':
			// Opening delimiter?
			if (index < final) && (content[index+1] == '%') {
				// Not escaped?
				if (index == 0) || (content[index-1] != '\\') {
					start = index
					index += 2
				}
			}

		case '%':
			// Closing delimiter?
			if (index < final) && (content[index+1] == '>') {
				// Not escaped?
				if (index == 0) || (content[index-1] != '\\') {
					index += 2
					if start != -1 {
						tags = append(tags, tag{start, index})
						start = -1
					} else {
						return nil, -1, fmt.Errorf("closing delimiter without an opening delimiter at position %d", index)
					}
				}
			}
		}
	}

	return tags, final, nil
}
