package commands

import (
	"github.com/spf13/cobra"
	kutiljs "github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/js"
	"github.com/tliron/prudence/platform"
)

var arguments map[string]string
var watch bool

func init() {
	rootCommand.AddCommand(runCommand)
	runCommand.Flags().StringToStringVarP(&arguments, "argument", "a", make(map[string]string), "arguments (format is name=value)")
	runCommand.Flags().BoolVarP(&watch, "watch", "w", true, "whether to watch dependent files and restart if they are changed")
}

var runCommand = &cobra.Command{
	Use:   "run [Script PATH or URL]",
	Short: "Run",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		util.OnExit(platform.Stop)

		context := urlpkg.NewContext()
		util.OnExit(func() {
			if err := context.Release(); err != nil {
				log.Errorf("%s", err.Error())
			}
		})

		environment := js.NewEnvironment(context)
		util.OnExit(func() {
			if err := environment.Release(); err != nil {
				log.Errorf("%s", err.Error())
			}
		})

		if watch {
			err := environment.Watch(func(id string, module *kutiljs.Module) {
				if module != nil {
					log.Infof("module changed: %s", module.Id)
				} else {
					log.Infof("file changed: %s", id)
				}
				platform.Restart()
			})
			if err != nil {
				log.Warningf("watch feature not supported on this platform")
			}
		}

		_, err := environment.RequireID(args[0])
		util.FailOnError(err)
	},
}
