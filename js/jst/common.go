package jst

import (
	"github.com/tliron/go-scriptlet/jst"
	"github.com/tliron/go-scriptlet/jst/sugar"
)

func RegisterDefaultSugar() {
	jst.RegisterSugar("#", sugar.HandleComment)
	jst.RegisterSugar("=", sugar.HandleExpression)
	jst.RegisterSugar("+", sugar.HandleInsert)
	jst.RegisterSugar("!", sugar.HandleCapture)
	jst.RegisterSugar("^", sugar.HandleRender)

	jst.RegisterSugar("&", HandleEmbed)
	jst.RegisterSugar("$", HandleSignature)
}
