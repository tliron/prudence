package platform

import (
	"fmt"
)

type RenderFunc func(content string, getRelativeURL GetRelativeURL) (string, error)

var renderers = make(map[string]RenderFunc)

func RegisterRenderer(renderer string, render RenderFunc) {
	renderers[renderer] = render
}

func GetRenderer(renderer string) (RenderFunc, error) {
	if renderer == "" {
		// Empty string means nil renderer
		return nil, nil
	} else if render, ok := renderers[renderer]; ok {
		return render, nil
	} else {
		return nil, fmt.Errorf("unsupported renderer: %s", renderer)
	}
}

func Render(content string, renderer string, getRelativeURL GetRelativeURL) (string, error) {
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