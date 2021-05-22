package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var plugins []string
var replace map[string]string
var version string
var executable string
var go_ string
var work string

func init() {
	rootCommand.AddCommand(buildCommand)
	buildCommand.Flags().StringArrayVarP(&plugins, "plugin", "p", nil, "plugin module (may include version after \"@\")")
	buildCommand.Flags().StringToStringVarP(&replace, "replace", "r", nil, "replace a module (format is module=path)")
	buildCommand.Flags().StringVarP(&version, "version", "e", "", "Prudence version (leave empty to use the latest version)")
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

func Build() {
	rootDirectory := GetWorkDirectory()

	sourceDirectory := filepath.Join(rootDirectory, executable)
	err := os.Mkdir(sourceDirectory, 0700)
	util.FailOnError(err)

	CreateMain(sourceDirectory)
	Command(rootDirectory, go_, "mod", "init", "github.com/tliron/prudence-x")
	FixGoMod(rootDirectory)

	if version != "" {
		log.Infof("getting Prudence version %q", version)
		Command(rootDirectory, go_, "get", "github.com/tliron/prudence@"+version)
	}

	for _, plugin := range plugins {
		plugin_ := strings.SplitN(plugin, "@", 2)
		if len(plugin_) > 1 {
			log.Infof("getting plugin %q version %q", plugin[0], plugin[1])
			Command(rootDirectory, go_, "get", plugin)
		}
	}

	Command(rootDirectory, go_, "mod", "tidy")
	Command(sourceDirectory, go_, "install", ".")

	gobin, err := util.GetGoBin()
	util.FailOnError(err)
	terminal.Printf("built: %s\n", filepath.Join(gobin, executable))
}

func GetWorkDirectory() string {
	if work != "" {
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

func CreateMain(dir string) {
	name := filepath.Join(dir, "main.go")
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0600)
	util.FailOnError(err)

	_, err = file.WriteString(main1)
	util.FailOnError(err)

	for _, plugin := range plugins {
		plugin_ := strings.SplitN(plugin, "@", 2)
		fmt.Fprintf(file, "\t_ %q\n", plugin_[0])
		log.Infof("added plugin %q", plugin_[0])
	}

	_, err = file.WriteString(main2)
	util.FailOnError(err)

	err = file.Close()
	util.FailOnError(err)
}

func FixGoMod(dir string) {
	if (replace != nil) && (len(replace) > 0) {
		name := filepath.Join(dir, "go.mod")
		file, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
		util.FailOnError(err)
		for module, path := range replace {
			_, err = file.WriteString(fmt.Sprintf("replace %s => %q\n", module, path))
			util.FailOnError(err)
			log.Infof("replaced module %q => %q", module, path)
		}
		err = file.Close()
		util.FailOnError(err)
	}
}

func Command(dir string, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	_, err := cmd.Output()
	FailOnCommandError(err)
}

func FailOnCommandError(err error) {
	if err_, ok := err.(*exec.ExitError); ok {
		util.Failf("%s\n%s", err_.Error(), util.BytesToString(err_.Stderr))
	} else {
		util.FailOnError(err)
	}
}
