package main

import (
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/prudence/commands"

	_ "github.com/tliron/commonlog/simple"
)

func main() {
	util.ExitOnSIGTERM()
	commands.Execute()
	util.Exit(0)
}
