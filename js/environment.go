package js

import (
	"github.com/tliron/exturl"
	"github.com/tliron/kutil/js"

	_ "github.com/tliron/prudence/distributed"
	_ "github.com/tliron/prudence/jst"
	_ "github.com/tliron/prudence/local"
	_ "github.com/tliron/prudence/memory"
	_ "github.com/tliron/prudence/render"
	_ "github.com/tliron/prudence/tiered"
)

func NewEnvironment(urlContext *exturl.Context, path []exturl.URL, arguments map[string]string) *js.Environment {
	environment := js.NewEnvironment(urlContext, path)
	environment.Extensions = newExtensions(arguments)
	environment.Precompile = precompile
	return environment
}
