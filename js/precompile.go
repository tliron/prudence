package js

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

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
			inputPath := fileUrl.Path
			outputPath := inputPath[:len(inputPath)-2] + "js"
			return tsc(inputPath, outputPath)
		} else {
			return "", errors.New("can only transpile local files")
		}

	/*case ".tsx", ".jsx":
	if fileUrl, ok := url.(*urlpkg.FileURL); ok {
		inputPath := fileUrl.Path
		outputPath := inputPath[:len(inputPath)-3] + "js"
		return tsc(inputPath, outputPath, "--jsx", "react")
	} else {
		return "", errors.New("can only transpile local files")
	}*/

	default:
		return script, nil
	}
}

var tscLock sync.Mutex

func tsc(inputPath string, outputPath string, args ...string) (string, error) {
	/*if _, err := os.Stat(outputPath); err == nil {
		return readFile(outputPath)
	} else if !os.IsNotExist(err) {
		return "", err
	}*/

	tscLock.Lock()
	defer tscLock.Unlock()

	log.Infof("tsc: %q to %q", inputPath, outputPath)
	args = append(args, "--target", "ES5", "--module", "commonjs", inputPath)
	cmd := exec.Command("tsc", args...)
	if stdout, err := cmd.Output(); err == nil {
		return readFile(outputPath)
	} else if err_, ok := err.(*exec.ExitError); ok {
		return "", fmt.Errorf("%s\n%s\n%s", err_.Error(), util.BytesToString(stdout), util.BytesToString(err_.Stderr))
	} else {
		return "", err
	}
}

func readFile(path string) (string, error) {
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
}
