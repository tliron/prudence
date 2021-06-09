package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var modules []string
var directories []string
var replacements map[string]string
var version string
var local string
var output string
var executable string
var go_ string
var work string

func init() {
	rootCommand.AddCommand(buildCommand)
	buildCommand.Flags().StringArrayVarP(&modules, "module", "m", nil, "add a module by name (may include version after \"@\")")
	buildCommand.Flags().StringArrayVarP(&directories, "directory", "d", nil, "add a module from a directory (must contain go.mod)")
	buildCommand.Flags().StringToStringVarP(&replacements, "replace", "r", make(map[string]string), "replace a module (format is module=path)")
	buildCommand.Flags().StringVarP(&version, "version", "e", "", "Prudence version (leave empty to use the latest version)")
	buildCommand.Flags().StringVarP(&local, "local", "c", "", "path to local Prudence source")
	buildCommand.Flags().StringVarP(&output, "output", "o", "", "output directory (defaults to $GOBIN or $GOPATH/bin or $HOME/go/bin)")
	buildCommand.Flags().StringVarP(&executable, "executable", "x", "prudence", "Prudence executable name")
	buildCommand.Flags().StringVarP(&go_, "go", "g", "go", "go binary")
	buildCommand.Flags().StringVarP(&work, "work", "w", "", "work directory (leave empty to create a temporary directory)")
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Build a custom Prudence executable",
	Run: func(cmd *cobra.Command, args []string) {
		Build()
	},
}

var main1 = `package main

import (
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/prudence/commands"

	_ "github.com/tliron/kutil/logging/simple"
`

var main2 = `)

func main() {
	commands.Execute()
	util.Exit(0)
}
`

var pluginPrefix = "prudence-x-"

func Build() {
	rootDirectory := GetWorkDirectory()

	sourceDirectory := filepath.Join(rootDirectory, executable)
	err := os.Mkdir(sourceDirectory, 0700)
	util.FailOnError(err)

	ConvertDirectories()
	CreateMain(sourceDirectory)
	Command(rootDirectory, nil, go_, "mod", "init", "github.com/tliron/prudence-x")

	if local != "" {
		replacements["github.com/tliron/prudence"] = local
	} else if version != "" {
		log.Infof("getting Prudence version %q", version)
		Command(rootDirectory, nil, go_, "get", "github.com/tliron/prudence@"+version)
	} else {
		log.Info("getting latest version of Prudence")
		Command(rootDirectory, nil, go_, "get", "github.com/tliron/prudence")
		version = Command(rootDirectory, nil, go_, "list", "-m", "-f", "{{ .Version }}", "github.com/tliron/prudence")
		version = strings.TrimSpace(version)
	}

	FixGoMod(rootDirectory)

	for _, plugin := range modules {
		plugin_ := strings.SplitN(plugin, "@", 2)
		if len(plugin_) > 1 {
			log.Infof("getting plugin %q version %q", plugin[0], plugin[1])
			Command(rootDirectory, nil, go_, "get", plugin)
		}
	}

	Command(rootDirectory, nil, go_, "mod", "tidy")

	if output == "" {
		output, err = util.GetGoBin()
		util.FailOnError(err)
	}

	output, err = filepath.Abs(output)
	util.FailOnError(err)

	path := filepath.Join(output, executable)

	timestamp := time.Now().Format("2006-01-02 15:04:05 MST")

	// Note: When "local" is set, the version will *not* be automatically figured out
	// and any "version" provided will be used as is without validation
	version_ := version
	if version_ == "" {
		version_ = "custom"
	} else {
		version_ += "-custom"
	}

	ldflags := fmt.Sprintf("-X 'github.com/tliron/kutil/version.GitVersion=%s' -X 'github.com/tliron/kutil/version.Timestamp=%s'", version_, timestamp)

	log.Infof("building: %s version %s", path, version_)
	Command(sourceDirectory, []string{"GOBIN=" + output}, go_, "install", "-ldflags", ldflags, ".")
	terminal.Printf("built: %s\n", path)
}

func GetWorkDirectory() string {
	if work != "" {
		var err error
		work, err = filepath.Abs(work)
		util.FailOnError(err)
		log.Infof("using work directory %q", work)
		return work
	}

	directory, err := os.MkdirTemp("", "xprudence-")
	util.FailOnError(err)
	log.Infof("created work directory %q", directory)
	util.OnExit(func() {
		os.RemoveAll(directory)
	})
	return directory
}

func ConvertDirectories() {
	var index int64
	var err error
	for _, directory := range directories {
		directory, err = filepath.Abs(directory)
		util.FailOnError(err)
		name := pluginPrefix + strconv.FormatInt(index, 10)
		modules = append(modules, name)
		replacements[name] = directory
		index++
	}
}

func CreateMain(dir string) {
	name := filepath.Join(dir, "main.go")
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0600)
	util.FailOnError(err)

	_, err = file.WriteString(main1)
	util.FailOnError(err)

	for _, module := range modules {
		module_ := strings.SplitN(module, "@", 2)
		fmt.Fprintf(file, "\t_ %q\n", module_[0])
		log.Infof("added module %q", module_[0])
	}

	_, err = file.WriteString(main2)
	util.FailOnError(err)

	err = file.Close()
	util.FailOnError(err)
}

func FixGoMod(dir string) {
	if len(replacements) > 0 {
		name := filepath.Join(dir, "go.mod")
		file, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
		util.FailOnError(err)

		for module, path := range replacements {
			_, err = file.WriteString(fmt.Sprintf("replace %s => %q\n", module, path))
			util.FailOnError(err)
			log.Infof("replaced module %q => %q", module, path)
		}

		err = file.Close()
		util.FailOnError(err)
	}
}

func Command(dir string, env []string, name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	output, err := cmd.Output()
	FailOnCommandError(err)
	return util.BytesToString(output)
}

func FailOnCommandError(err error) {
	if err_, ok := err.(*exec.ExitError); ok {
		util.Failf("%s\n%s", err_.Error(), util.BytesToString(err_.Stderr))
	} else {
		util.FailOnError(err)
	}
}
