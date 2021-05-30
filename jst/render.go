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
	var context platform.Context

	context.Builder.WriteString("exports.present = function(context) {\n")

	last := 0

	if matches := jstRe.FindAllStringSubmatchIndex(content, -1); matches != nil {
		for _, match := range matches {
			start := match[0]
			end := match[1]

			//log.Debugf("match: %s", content[match[0]:match[1]])
			context.WriteLiteral(content[last:start])

			code := content[start+2 : end-2]
			last = end

			if code == "" {
				continue
			}

			// Skip trailing newline by default
			skipTrailingNewline := true

			if content[end-3] == '/' {
				// Explicitly allow trailing newline
				code = code[:len(code)-1]
				skipTrailingNewline = false
			}

			// Encode
			encoded := false
			platform.OnTags(func(prefix string, encodeTag platform.EncodeTagFunc) bool {
				if strings.HasPrefix(code, prefix) {
					if encodeTag(&context, code) {
						skipTrailingNewline = false
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

			if skipTrailingNewline {
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
