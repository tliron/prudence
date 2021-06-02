package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("$", EncodeSignature)
}

// platform.EncodeTagFunc signature
func EncodeSignature(context *platform.JSTContext, code string) bool {
	code = code[1:]

	if code == "$" {
		// End signature
		context.Builder.WriteString("context.endSignature();\n")
	} else {
		// Start signature
		context.Builder.WriteString("context.startSignature();\n")
		if weak := strings.Trim(code, " \n"); weak != "" {
			context.Builder.WriteString("if (")
			context.Builder.WriteString(weak)
			context.Builder.WriteString(") context.weakSignature = true;\n")
		}
	}

	return false
}
