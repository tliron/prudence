package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("^", HandleRender)
}

// platform.HandleTagFunc signature
func HandleRender(context *platform.JSTContext, code string) bool {
	code = code[1:]

	if code == "^" {
		// End render
		context.Builder.WriteString("context.endRender();\n")
	} else {
		// Start render
		context.Builder.WriteString("context.startRender(")
		context.Builder.WriteString(strings.TrimSpace(code))
		context.Builder.WriteString(", prudence.jsContext);\n")
	}

	return false
}
