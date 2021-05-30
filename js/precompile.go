package js

import (
	"strings"

	"github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/platform"
)

// js.PrecompileFunc signature
func precompile(url urlpkg.URL, script string, context *js.Context) (string, error) {
	if strings.HasSuffix(url.String(), ".jst") {
		if script_, err := platform.Render(script, "jst", context.Resolve); err == nil {
			return script_, nil
		} else {
			return "", err
		}
	} else {
		return script, nil
	}
}
