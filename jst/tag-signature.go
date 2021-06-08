package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("$", HandleSignature)
}

// platform.HandleTagFunc signature
func HandleSignature(context *platform.JSTContext, code string) bool {
	code = code[1:]

	if code == "$" {
		// End signature
		context.Builder.WriteString("context.endSignature();\n")
	} else {
		// Start signature
		context.Builder.WriteString("context.startSignature();\n")
		if weak := strings.TrimSpace(code); weak != "" {
			context.Builder.WriteString("if (")
			context.Builder.WriteString(weak)
			context.Builder.WriteString(") context.response.weakSignature = true;\n")
		}
	}

	return false
}
