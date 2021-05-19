package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("&", EncodeEmbed)
}

// platform.EncodeTagFunc signature
func EncodeEmbed(context *platform.Context, code string) bool {
	code = code[1:]
	suffix := context.NextSuffix()

	context.Builder.WriteString("const __args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = [")
	context.Builder.WriteString(strings.Trim(code, " \n"))
	context.Builder.WriteString("];\n")

	context.Builder.WriteString("const __hook")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = prudence.hook(__args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString("[0], 'present');\n")

	context.Builder.WriteString("const __context")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = context.copy();\n")

	context.Builder.WriteString("__context")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(".cacheKey = __args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString("[1] || (context.cacheKey + '|")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString("');\n")

	context.Builder.WriteString("__context")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(".embed(__hook")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(");\n")

	return false
}
