package render

import (
	"github.com/gomarkdown/markdown"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterRenderer("markdown", RenderMarkdown)
	platform.RegisterRenderer("md", RenderMarkdown)
}

// platform.RenderFunc signature
func RenderMarkdown(content string, resolve js.ResolveFunc) (string, error) {
	return util.BytesToString(markdown.ToHTML(util.StringToBytes(content), nil, nil)), nil
}
