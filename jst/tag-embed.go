package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("&", HandleEmbed)
}

// platform.HandleTagFunc signature
func HandleEmbed(context *platform.JSTContext, code string) bool {
	code = code[1:]
	suffix := context.NextSuffix()

	context.Builder.WriteString("const __args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = [")
	context.Builder.WriteString(strings.TrimSpace(code))
	context.Builder.WriteString("];\n")

	context.Builder.WriteString("const __present")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = require(__args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString("[0]).present;\n")

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
	context.Builder.WriteString(".embed(__present")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(");\n")

	return false
}
