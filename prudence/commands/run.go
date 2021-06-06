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

/*
type X struct{}

func (self X) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	path := filepath.Join("examples/hello-world/myapp/static", request.URL.Path)
	http.ServeFile(responseWriter, request, path)
}
*/

var runCommand = &cobra.Command{
	Use:   "run [Script PATH or URL]",
	Short: "Run",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		util.OnExit(platform.Stop)

		/*
			server := http.Server{
				Addr:         "localhost:8080",
				ReadTimeout:  time.Duration(time.Second * 5),
				WriteTimeout: time.Duration(time.Second * 5),
				Handler:      X{},
			}
			listener, err := net.Listen("tcp", "localhost:8080")
			util.FailOnError(err)
			err = server.Serve(listener)
			util.FailOnError(err)
		*/

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

		restart := func(id string, module *kutiljs.Module) {
			if module != nil {
				log.Infof("module changed: %s", module.Id)
			} else if id != "" {
				log.Infof("file changed: %s", id)
			}

			environment.ClearCache()
			_, err := environment.RequireID(args[0])
			util.FailOnError(err)
		}

		if watch {
			if err := environment.Watch(restart); err != nil {
				log.Warningf("watch feature not supported on this platform")
			}
		}

		restart("", nil)

		// Block forever
		<-make(chan bool, 0)
	},
}
