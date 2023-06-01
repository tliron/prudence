package js

import (
	"path/filepath"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/exturl"
	"github.com/tliron/prudence/platform"
)

// js.PrecompileFunc signature
func precompile(url exturl.URL, script string, context *commonjs.Context) (string, error) {
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
