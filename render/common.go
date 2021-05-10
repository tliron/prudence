package render

import (
	"fmt"

	"github.com/tliron/kutil/logging"
	"github.com/tliron/prudence/js/common"
)

var log = logging.GetLogger("prudence.render")

type RenderFunc func(content string, getRelativeURL common.GetRelativeURL) (string, error)

var renderFuncs = make(map[string]RenderFunc)

func Register(renderer string, renderFunc RenderFunc) {
	renderFuncs[renderer] = renderFunc
}

func Render(content string, renderer string, getRelativeURL common.GetRelativeURL) (string, error) {
	if render, ok := renderFuncs[renderer]; ok {
		return render(content, getRelativeURL)
	} else {
		return "", fmt.Errorf("unsupported renderer: %s", renderer)
	}
}
