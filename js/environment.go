package js

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/commonlog"
	"github.com/tliron/exturl"
	"github.com/tliron/go-scriptlet/jst"
	"github.com/tliron/go-scriptlet/markdown"
	"github.com/tliron/go-scriptlet/minify"
	prudencejst "github.com/tliron/prudence/js/jst"
	"github.com/tliron/prudence/platform"
)

func init() {
	jst.RegisterDefaultRenderers()
	markdown.RegisterDefaultRenderers()
	minify.RegisterDefaultRenderers()

	prudencejst.RegisterDefaultSugar()
}

var log = commonlog.GetLogger("prudence.js")

func NewEnvironment(arguments map[string]string, urlContext *exturl.Context, basePaths ...exturl.URL) *commonjs.Environment {
	environment := jst.NewDefaultEnvironment(log, urlContext, basePaths...)

	environment.Extensions = append(environment.Extensions, commonjs.Extension{
		Name:   "prudence",
		Create: CreatePrudenceExtension,
	})

	environment.Extensions = append(environment.Extensions, commonjs.NewExtensions(platform.Extensions)...)

	return environment
}
