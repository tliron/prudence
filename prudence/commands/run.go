package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/js"
)

func init() {
	rootCommand.AddCommand(runCommand)
}

var runCommand = &cobra.Command{
	Use:   "run [Script PATH or URL]",
	Short: "Run",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := js.NewAPI(nil).Import(args[0])
		util.FailOnError(err)
	},
}
