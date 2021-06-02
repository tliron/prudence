package jst

import (
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterTag("#", EncodeComment)
}

// platform.EncodeTagFunc signature
func EncodeComment(context *platform.JSTContext, code string) bool {
	return false
}
