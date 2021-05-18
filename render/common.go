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

func GetRenderer(renderer string) (RenderFunc, error) {
	if renderer == "" {
		// Empty string means nil renderer
		return nil, nil
	} else if render, ok := renderFuncs[renderer]; ok {
		return render, nil
	} else {
		return nil, fmt.Errorf("unsupported renderer: %s", renderer)
	}
}

func Render(content string, renderer string, getRelativeURL common.GetRelativeURL) (string, error) {
	if render, err := GetRenderer(renderer); err == nil {
		if render == nil {
			// Renderer can be nil
			return content, nil
		} else {
			return render(content, getRelativeURL)
		}
	} else {
		return "", err
	}
}
