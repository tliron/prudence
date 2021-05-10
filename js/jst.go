package js

import (
	"regexp"
	"strconv"
	"strings"
)

var jstRe = regexp.MustCompile(`(?s)<%.*?%>`)

func RenderJST(content string) (string, error) {
	var builder strings.Builder

	builder.WriteString("function represent(context) {\n")

	last := 0
	var representers []string

	if matches := jstRe.FindAllStringSubmatchIndex(content, -1); matches != nil {
		for _, match := range matches {
			start := match[0]
			end := match[1]

			//log.Infof("match: %s", content[match[0]:match[1]])
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
					// Load

					// Render?
					var renderer string
					if code[1] != ' ' {
						code = code[1:]
						if space := strings.IndexRune(code, ' '); space != -1 {
							renderer = code[:space]
							code = code[space+1:]
						}
					} else {
						code = code[1:]
					}

					if renderer != "" {
						builder.WriteString("context.write(prudence.render(prudence.load(")
						builder.WriteString(strings.Trim(code, " \n"))
						builder.WriteString("), '")
						builder.WriteString(renderer)
						builder.WriteString("'));\n")
					} else {
						builder.WriteString("context.write(prudence.load(")
						builder.WriteString(strings.Trim(code, " \n"))
						builder.WriteString("));\n")
					}

				case '&':
					// Represent
					code = code[1:]
					builder.WriteString("__representer")
					builder.WriteString(strconv.FormatInt(int64(len(representers)), 10))
					builder.WriteString(".callable(null, context);\n")

					representers = append(representers, strings.Trim(code, " \n"))

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

	// Hook representers globally
	for index, representer := range representers {
		builder.WriteString("var __representer")
		builder.WriteString(strconv.FormatInt(int64(index), 10))
		builder.WriteString(" = prudence.hook(")
		builder.WriteString(representer)
		builder.WriteString(", 'represent');")
	}

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
