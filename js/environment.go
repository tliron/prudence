package js

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/exturl"

	_ "github.com/tliron/prudence/distributed"
	_ "github.com/tliron/prudence/jst"
	_ "github.com/tliron/prudence/local"
	_ "github.com/tliron/prudence/memory"
	_ "github.com/tliron/prudence/render"
	_ "github.com/tliron/prudence/tiered"
)

func NewEnvironment(urlContext *exturl.Context, path []exturl.URL, arguments map[string]string) *commonjs.Environment {
	environment := commonjs.NewEnvironment(urlContext, path)
	environment.Extensions = newExtensions(arguments)
	environment.Precompile = precompile
	return environment
}
