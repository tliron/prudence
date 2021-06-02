package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("*", EncodeCacheDuration)
}

// platform.EncodeTagFunc signature
func EncodeCacheDuration(context *platform.JSTContext, code string) bool {
	code = code[1:]

	context.Builder.WriteString("context.cacheDuration = ")
	context.Builder.WriteString(strings.Trim(code, " \n"))
	context.Builder.WriteString(";\n")

	return false
}
