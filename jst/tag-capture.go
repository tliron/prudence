package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("!", HandleCapture)
}

// platform.HandleTagFunc signature
func HandleCapture(context *platform.JSTContext, code string) bool {
	code = code[1:]

	if code == "!" {
		// End render
		context.Builder.WriteString("context.endCapture();\n")
	} else {
		// Start render
		context.Builder.WriteString("context.startCapture(")
		context.Builder.WriteString(strings.Trim(code, " \n"))
		context.Builder.WriteString(");\n")
	}

	return false
}
