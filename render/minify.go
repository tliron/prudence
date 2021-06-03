package render

import (
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
	kutiljs "github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

var minify_ *minify.M

func init() {
	minify_ = minify.New()
	minify_.AddFunc("text/css", css.Minify)
	minify_.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
	})
	minify_.AddFunc("image/svg+xml", svg.Minify)
	minify_.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	minify_.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	minify_.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	platform.RegisterRenderer("mincss", RenderMinifyCSS)
	platform.RegisterRenderer("minhtml", RenderMinifyHTML)
	platform.RegisterRenderer("minsvg", RenderMinifySVG)
	platform.RegisterRenderer("minjs", RenderMinifyJavaScript)
	platform.RegisterRenderer("minjson", RenderMinifyJSON)
	platform.RegisterRenderer("minxml", RenderMinifyXML)
}

// platform.RenderFunc signature
func RenderMinifyCSS(content string, context *kutiljs.Context) (string, error) {
	return minify_.String("text/css", content)
}

// platform.RenderFunc signature
func RenderMinifyHTML(content string, context *kutiljs.Context) (string, error) {
	return minify_.String("text/html", content)
}

// platform.RenderFunc signature
func RenderMinifySVG(content string, context *kutiljs.Context) (string, error) {
	return minify_.String("image/svg+xml", content)
}

// platform.RenderFunc signature
func RenderMinifyJavaScript(content string, context *kutiljs.Context) (string, error) {
	return minify_.String("text/javascript", content)
}

// platform.RenderFunc signature
func RenderMinifyJSON(content string, context *kutiljs.Context) (string, error) {
	return minify_.String("application/json", content)
}

// platform.RenderFunc signature
func RenderMinifyXML(content string, context *kutiljs.Context) (string, error) {
	return minify_.String("application/xml", content)
}
