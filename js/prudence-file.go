package js

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/tliron/kutil/util"
)

func (self *PrudenceAPI) Exec(name string, arguments ...string) (string, error) {
	cmd := exec.Command(name, arguments...)
	if out, err := cmd.Output(); err == nil {
		return util.BytesToString(out), nil
	} else if err_, ok := err.(*exec.ExitError); ok {
		return "", fmt.Errorf("%s\n%s", err_.Error(), util.BytesToString(err_.Stderr))
	} else {
		return "", err
	}
}

func (self *PrudenceAPI) TemporaryFile(pattern string, directory string) (string, error) {
	if file, err := os.CreateTemp(directory, pattern); err == nil {
		name := file.Name()
		os.Remove(name)
		return name, nil
	} else {
		return "", err
	}
}

func (self *PrudenceAPI) TemporaryDirectory(pattern string, directory string) (string, error) {
	return os.MkdirTemp(directory, pattern)
}
