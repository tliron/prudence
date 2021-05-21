package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/js"
)

var arguments map[string]string

func init() {
	rootCommand.AddCommand(runCommand)
	runCommand.Flags().StringToStringVarP(&arguments, "argument", "a", make(map[string]string), "arguments (format is name=value)")
}

var runCommand = &cobra.Command{
	Use:   "run [Script PATH or URL]",
	Short: "Run",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := js.NewPrudenceAPI(nil).Run(args[0])
		util.FailOnError(err)
	},
}
