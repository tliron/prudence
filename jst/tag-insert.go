package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("+", HandleInsert)
}

// platform.HandleTagFunc signature
func HandleInsert(context *platform.JSTContext, code string) bool {
	code = code[1:]
	suffix := context.NextSuffix()

	context.Builder.WriteString("const __args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = [")
	context.Builder.WriteString(strings.TrimSpace(code))
	context.Builder.WriteString("];\n")

	context.Builder.WriteString("var __insert")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = prudence.loadString(__args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString("[0]);\n")

	context.Builder.WriteString("if (__args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(".length > 1) ")
	context.Builder.WriteString("__insert")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(" = prudence.render(__insert")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(", __args")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString("[1]);\n")

	context.Builder.WriteString("context.write(__insert")
	context.Builder.WriteString(suffix)
	context.Builder.WriteString(");\n")

	return false
}
