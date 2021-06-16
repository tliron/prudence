package commands

import (
	"github.com/spf13/cobra"
	kutiljs "github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/kutil/version"
	"github.com/tliron/prudence/js"
	"github.com/tliron/prudence/platform"
)

var arguments map[string]string
var watch bool

func init() {
	rootCommand.AddCommand(runCommand)
	runCommand.Flags().StringToStringVarP(&arguments, "argument", "a", make(map[string]string), "arguments (format is name=value)")
	runCommand.Flags().BoolVarP(&watch, "watch", "w", true, "whether to watch dependent files and restart if they are changed")
	runCommand.Flags().StringVarP(&platform.NCSAFilename, "ncsa", "n", "", "NCSA log filename (or special values \"stdout\" and \"stderr\")")
}

var runCommand = &cobra.Command{
	Use:   "run [Script PATH or URL]",
	Short: "Run",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startId := args[0]

		util.OnExit(platform.Stop)

		urlContext := urlpkg.NewContext()
		util.OnExit(func() {
			if err := urlContext.Release(); err != nil {
				log.Errorf("%s", err.Error())
			}
		})

		environment := js.NewEnvironment(urlContext, arguments)
		util.OnExit(func() {
			if err := environment.Release(); err != nil {
				log.Errorf("%s", err.Error())
			}
		})

		restart := func(id string, module *kutiljs.Module) {
			if module != nil {
				log.Infof("module changed: %s", module.Id)
			} else if id != "" {
				log.Infof("file changed: %s", id)
			}

			environment.ClearCache()

			environment.Lock.Lock()
			_, err := environment.RequireID(startId)
			environment.Lock.Unlock()

			util.FailOnError(err)
		}

		if watch {
			if err := environment.Watch(restart); err != nil {
				log.Warningf("watch feature not supported on this platform")
			}
		}

		log.Noticef("Prudence version %s", version.GitVersion)

		restart("", nil)

		// Block forever
		<-make(chan bool, 0)
	},
}
