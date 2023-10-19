package jst

import (
	"strings"

	"github.com/tliron/go-scriptlet/jst"
)

// ([jst.HandleSugarFunc] signature)
func HandleSignature(scriptletContext *jst.ScriptletContext, prefix string, code string) (bool, error) {
	code = code[len(prefix):]

	if code == prefix {
		// End signature
		return false, scriptletContext.WriteString("this.endSignature();\n")
	} else {
		// Start signature
		if err := scriptletContext.WriteString("this.startSignature();\n"); err != nil {
			return false, err
		}
		if weak := strings.TrimSpace(code); weak != "" {
			if err := scriptletContext.WriteString("if ("); err != nil {
				return false, err
			}
			if err := scriptletContext.WriteString(weak); err != nil {
				return false, err
			}
			return false, scriptletContext.WriteString(") this.response.weakSignature = true;\n")
		}
		return false, nil
	}
}
