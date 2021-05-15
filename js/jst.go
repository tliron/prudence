package js

import (
	"regexp"
	"strconv"
	"strings"
)

var jstRe = regexp.MustCompile(`(?s)<%.*?%>`)

func RenderJST(content string) (string, error) {
	var builder strings.Builder

	builder.WriteString("function present(context) {\n")

	last := 0
	var embedIndex int64
	//var presenters []string

	if matches := jstRe.FindAllStringSubmatchIndex(content, -1); matches != nil {
		for _, match := range matches {
			start := match[0]
			end := match[1]

			//log.Debugf("match: %s", content[match[0]:match[1]])
			writeLiteral(&builder, content[last:start])

			code := content[start+2 : end-2]
			last = end

			if code != "" {
				switch code[0] {
				case '=':
					// Don't ignore trailing newline

				default:
					// Ignore trailing newline
					if content[end] == '\n' {
						last++
					}
				}

				switch code[0] {
				case '#':
					// Comment, do nothing

				case '=':
					// Write expression
					code = code[1:]
					builder.WriteString("context.write(String(")
					builder.WriteString(strings.Trim(code, " \n"))
					builder.WriteString("));\n")

				case '+':
					// Insert

					code = code[1:]
					suffix := strconv.FormatInt(embedIndex, 10)
					embedIndex++

					builder.WriteString("const __args")
					builder.WriteString(suffix)
					builder.WriteString(" = [")
					builder.WriteString(strings.Trim(code, " \n"))
					builder.WriteString("];\n")

					builder.WriteString("var __insert")
					builder.WriteString(suffix)
					builder.WriteString(" = prudence.load(__args")
					builder.WriteString(suffix)
					builder.WriteString("[0]);\n")

					builder.WriteString("if (__args")
					builder.WriteString(suffix)
					builder.WriteString(".length > 1) ")
					builder.WriteString("__insert")
					builder.WriteString(suffix)
					builder.WriteString(" = prudence.render(__insert")
					builder.WriteString(suffix)
					builder.WriteString(", __args")
					builder.WriteString(suffix)
					builder.WriteString("[1]);\n")

					builder.WriteString("context.write(__insert")
					builder.WriteString(suffix)
					builder.WriteString(");\n")

				case '&':
					// Embed

					code = code[1:]
					suffix := strconv.FormatInt(embedIndex, 10)
					embedIndex++

					builder.WriteString("const __args")
					builder.WriteString(suffix)
					builder.WriteString(" = [")
					builder.WriteString(strings.Trim(code, " \n"))
					builder.WriteString("];\n")

					builder.WriteString("const __hook")
					builder.WriteString(suffix)
					builder.WriteString(" = prudence.hook(__args")
					builder.WriteString(suffix)
					builder.WriteString("[0], 'present');\n")

					builder.WriteString("const __context")
					builder.WriteString(suffix)
					builder.WriteString(" = context.copy();\n")

					builder.WriteString("__context")
					builder.WriteString(suffix)
					builder.WriteString(".cacheKey = __args")
					builder.WriteString(suffix)
					builder.WriteString("[1] || (context.cacheKey + '|")
					builder.WriteString(suffix)
					builder.WriteString("');\n")

					builder.WriteString("__context")
					builder.WriteString(suffix)
					builder.WriteString(".embed(__hook")
					builder.WriteString(suffix)
					builder.WriteString(");\n")

				default:
					// As is
					builder.WriteString(strings.Trim(code, " \n"))
					builder.WriteRune('\n')
				}
			}
		}
	}

	writeLiteral(&builder, content[last:])
	builder.WriteString("}\n")

	log.Debugf("%s", builder.String())

	return builder.String(), nil
}

func writeLiteral(builder *strings.Builder, literal string) {
	if literal != "" {
		builder.WriteString("context.write('")
		for _, rune_ := range literal {
			switch rune_ {
			case '\n':
				builder.WriteString("\\n")
			case '\'':
				builder.WriteString("\\'")
			default:
				builder.WriteRune(rune_)
			}
		}
		builder.WriteString("');\n")
	}
}
