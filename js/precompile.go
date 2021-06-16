package js

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
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

	case ".ts":
		if fileUrl, ok := url.(*urlpkg.FileURL); ok {
			cmd := exec.Command("tsc", "--target", "ES5", fileUrl.Path)
			if stdout, err := cmd.Output(); err == nil {
				path := fileUrl.Path[:len(fileUrl.Path)-3] + ".js"
				if reader, err := os.Open(path); err == nil {
					defer reader.Close()
					if bytes, err := io.ReadAll(reader); err == nil {
						return util.BytesToString(bytes), nil
					} else {
						return "", err
					}
				} else {
					return "", err
				}
			} else if err_, ok := err.(*exec.ExitError); ok {
				return "", fmt.Errorf("%s\n%s\n%s", err_.Error(), util.BytesToString(stdout), util.BytesToString(err_.Stderr))
			} else {
				return "", err
			}
		} else {
			return "", errors.New("can only transpile TypeScript for local files")
		}

	default:
		return script, nil
	}
}
