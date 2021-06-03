package jst

import (
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("#", HandleComment)
}

// platform.HandleTagFunc signature
func HandleComment(context *platform.JSTContext, code string) bool {
	return false
}
