package jst

import (
	"strings"

	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("*", HandleCacheDuration)
}

// platform.HandleTagFunc signature
func HandleCacheDuration(context *platform.JSTContext, code string) bool {
	code = code[1:]

	context.Builder.WriteString("context.cacheDuration = ")
	context.Builder.WriteString(strings.TrimSpace(code))
	context.Builder.WriteString(";\n")

	return false
}
