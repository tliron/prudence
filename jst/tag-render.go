package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("^", EncodeRender)
}

// platform.EncodeTagFunc signature
func EncodeRender(context *platform.JSTContext, code string) bool {
	code = code[1:]

	if code == "^" {
		// End render
		context.Builder.WriteString("context.endRender();\n")
	} else {
		// Start render
		context.Builder.WriteString("context.startRender(")
		context.Builder.WriteString(strings.Trim(code, " \n"))
		context.Builder.WriteString(", prudence.resolve);\n")
	}

	return false
}
