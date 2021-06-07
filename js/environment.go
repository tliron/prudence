package js

import (
	"github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"

	_ "github.com/tliron/prudence/jst"
	_ "github.com/tliron/prudence/memory"
	_ "github.com/tliron/prudence/render"
)

func NewEnvironment(urlContext *urlpkg.Context, arguments map[string]string) *js.Environment {
	environment := js.NewEnvironment(urlContext)
	environment.Extensions = newExtensions(arguments)
	environment.Precompile = precompile
	return environment
}
