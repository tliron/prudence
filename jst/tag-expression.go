package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("=", HandleExpression)
}

// platform.HandleTagFunc signature
func HandleExpression(context *platform.JSTContext, code string) bool {
	code = code[1:]

	if (len(code) > 0) && (code[0] == '=') {
		// Variable
		context.Builder.WriteString("context.write(String(context.variables[")
		context.Builder.WriteString(strings.Trim(code[1:], " \n"))
		context.Builder.WriteString("]));\n")
	} else {
		// Expression
		context.Builder.WriteString("context.write(String(")
		context.Builder.WriteString(strings.Trim(code, " \n"))
		context.Builder.WriteString("));\n")
	}

	return true
}
