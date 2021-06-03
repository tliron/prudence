package jst

import (
	"regexp"
	"strings"

	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterRenderer("jst", RenderJST)
}

var jstRe = regexp.MustCompile(`(?s)<%.*?%>`)

// platform.RenderFunc signature
func RenderJST(content string, resolve js.ResolveFunc) (string, error) {
	var context platform.JSTContext

	context.Builder.WriteString("exports.present = function(context) {\n")

	last := 0

	// Escape
	content = strings.ReplaceAll(content, "\\<%", "<% context.writeString('<%'); %>")
	content = strings.ReplaceAll(content, "\\%>", "<% context.writeString('%>'); %>")

	if matches := jstRe.FindAllStringSubmatchIndex(content, -1); matches != nil {
		for _, match := range matches {
			start := match[0]
			end := match[1]
			//log.Debugf("match: %s", content[start:end])

			// Write previous chunk
			context.WriteLiteral(content[last:start])
			last = end

			code := content[start+2 : end-2]

			if code == "" {
				continue
			}

			// Swallow trailing newline by default
			swallowTrailingNewline := true

			if content[end-3] == '/' {
				// Disable the swallowing of trailing newline
				code = code[:len(code)-1]
				swallowTrailingNewline = false
			}

			// Handle
			encoded := false
			platform.OnTags(func(prefix string, handleTag platform.HandleTagFunc) bool {
				if strings.HasPrefix(code, prefix) {
					if handleTag(&context, code) {
						swallowTrailingNewline = false
					}
					encoded = true
					return false
				}
				return true
			})

			if !encoded {
				// As is
				context.Builder.WriteString(strings.Trim(code, " \n"))
				context.Builder.WriteRune('\n')
			}

			if swallowTrailingNewline {
				// Skip trailing newline
				if content[end] == '\n' {
					last++
				}
			}
		}
	}

	context.WriteLiteral(content[last:])
	context.Builder.WriteString("};\n")

	string_ := context.Builder.String()
	//log.Debugf("%s", string_)

	return string_, nil
}
