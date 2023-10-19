package jst

import (
	"strings"

	"github.com/tliron/go-scriptlet/jst"
)

// ([jst.HandleSugarFunc] signature)
func HandleEmbed(scriptletContext *jst.ScriptletContext, prefix string, code string) (bool, error) {
	prefixLength := len(prefix)
	code = code[prefixLength:]
	caching := strings.HasPrefix(code, prefix)
	suffix := scriptletContext.NextSuffix()

	if err := scriptletContext.WriteString("const __args"); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(suffix); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(" = ["); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(strings.TrimSpace(code)); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString("];\n"); err != nil {
		return false, err
	}

	if err := scriptletContext.WriteString("const __present"); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(suffix); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(" = bind(__args"); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(suffix); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString("[0], 'present');\n"); err != nil {
		return false, err
	}

	if err := scriptletContext.WriteString("const __context"); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(suffix); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(" = this.clone();\n"); err != nil {
		return false, err
	}

	if caching {
		code = code[prefixLength:]

		if err := scriptletContext.WriteString("__context"); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(suffix); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(".cacheDuration = __args"); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(suffix); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString("[0] || 1;\n"); err != nil {
			return false, err
		}

		if err := scriptletContext.WriteString("__context"); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(suffix); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(".cacheKey = __args"); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(suffix); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString("[2] || this.cacheKey + ';"); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(suffix); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString("';\n"); err != nil {
			return false, err
		}

		if err := scriptletContext.WriteString("__context"); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(suffix); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(".cacheGroups = __args"); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString(suffix); err != nil {
			return false, err
		}
		if err := scriptletContext.WriteString("[3] || [];\n"); err != nil {
			return false, err
		}
	}

	if err := scriptletContext.WriteString("__context"); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(suffix); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(".embed(__present"); err != nil {
		return false, err
	}
	if err := scriptletContext.WriteString(suffix); err != nil {
		return false, err
	}
	return false, scriptletContext.WriteString(", env.context);\n")
}
