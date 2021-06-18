package js

import (
	"path/filepath"

	"github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/platform"
)

// js.PrecompileFunc signature
func precompile(url urlpkg.URL, script string, context *js.Context) (string, error) {
	ext := filepath.Ext(url.String())
	switch ext {
	case ".jst":
		if script_, err := platform.Render(script, "jst", context); err == nil {
			return script_, nil
		} else {
			return "", err
		}

	default:
		return script, nil
	}
}
