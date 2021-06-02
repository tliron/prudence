package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("=", EncodeExpression)
}

// platform.EncodeTagFunc signature
func EncodeExpression(context *platform.JSTContext, code string) bool {
	code = code[1:]

	context.Builder.WriteString("context.write(String(")
	context.Builder.WriteString(strings.Trim(code, " \n"))
	context.Builder.WriteString("));\n")

	return true
}
