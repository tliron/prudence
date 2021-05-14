package render

import (
	"github.com/gomarkdown/markdown"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/js/common"
)

func init() {
	Register("markdown", RenderMarkdown)
	Register("md", RenderMarkdown)
}

// RenderFunc signature
func RenderMarkdown(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	return util.BytesToString(markdown.ToHTML(util.StringToBytes(content), nil, nil)), nil
}
