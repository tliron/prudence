package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var plugins []string
var replace map[string]string

func init() {
	rootCommand.AddCommand(buildCommand)
	buildCommand.Flags().StringArrayVarP(&plugins, "plugin", "p", nil, "plugin module")
	buildCommand.Flags().StringToStringVarP(&replace, "replace", "r", nil, "replace a module (format is module=path)")
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
	root := CreateWorkDirectory()

	prudence := filepath.Join(root, "prudence")
	err := os.Mkdir(prudence, 0700)
	util.FailOnError(err)

	CreateMain(prudence)
	Command(root, "go", "mod", "init", "github.com/tliron/prudence-x")
	FixGoMod(root)
	Command(root, "go", "mod", "tidy")
	Command(prudence, "go", "install", ".")
}

func CreateWorkDirectory() string {
	directory, err := os.MkdirTemp("", "xprudence-")
	util.FailOnError(err)
	log.Infof("created work directory: %s", directory)
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
		fmt.Fprintf(file, "\t_ %q\n", plugin)
		log.Infof("added plugin: %s", plugin)
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
			log.Infof("replaced module: %s => %q", module, path)
			_, err = file.WriteString(fmt.Sprintf("replace %s => %q\n", module, path))
			util.FailOnError(err)
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
