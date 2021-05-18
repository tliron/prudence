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
	"github.com/tliron/prudence/js/common"
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

	Register("mincss", RenderMinifyCSS)
	Register("minhtml", RenderMinifyHTML)
	Register("minsvg", RenderMinifySVG)
	Register("minjs", RenderMinifyJavaScript)
	Register("minjson", RenderMinifyJSON)
	Register("minxml", RenderMinifyXML)
}

// RenderFunc signature
func RenderMinifyCSS(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	return minify_.String("text/css", content)
}

// RenderFunc signature
func RenderMinifyHTML(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	return minify_.String("text/html", content)
}

// RenderFunc signature
func RenderMinifySVG(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	return minify_.String("image/svg+xml", content)
}

// RenderFunc signature
func RenderMinifyJavaScript(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	return minify_.String("text/javascript", content)
}

// RenderFunc signature
func RenderMinifyJSON(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	return minify_.String("application/json", content)
}

// RenderFunc signature
func RenderMinifyXML(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	return minify_.String("application/xml", content)
}
