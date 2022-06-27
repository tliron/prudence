package commands

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	kutiljs "github.com/tliron/kutil/js"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
	"github.com/tliron/kutil/version"
	"github.com/tliron/prudence/js"
	"github.com/tliron/prudence/platform"
)

var paths []string
var typescript string
var arguments map[string]string
var watch bool

func init() {
	rootCommand.AddCommand(runCommand)
	runCommand.Flags().StringArrayVarP(&paths, "path", "p", nil, "library path (appended after PRUDENCE_PATH environment variable)")
	runCommand.Flags().StringVarP(&typescript, "typescript", "t", "", "TypeScript project path (must have a tsconfig.json file)")
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
		util.OnExitError(urlContext.Release)

		var path_ []urlpkg.URL

		parsePaths := func(paths []string) {
			for _, path := range paths {
				if !strings.HasSuffix(path, "/") {
					path += "/"
				}
				pathUrl, err := urlpkg.NewValidURL(path, nil, urlContext)
				log.Infof("library path: %s", pathUrl.String())
				util.FailOnError(err)
				path_ = append(path_, pathUrl)
			}
		}

		parsePaths(filepath.SplitList(os.Getenv("PRUDENCE_PATH")))
		parsePaths(paths)

		environment := js.NewEnvironment(urlContext, path_, arguments)
		util.OnExitError(environment.Release)

		log.Noticef("Prudence version: %s", version.GitVersion)

		if typescript != "" {
			transpileTypeScript()
		}

		environment.OnChanged = func(id string, module *kutiljs.Module) {
			if module != nil {
				log.Infof("module changed: %s", module.Id)
			} else if id != "" {
				log.Infof("file changed: %s", id)
			}

			environment.Lock.Lock()

			if watch {
				if err := environment.RestartWatcher(); err != nil {
					log.Warningf("watch feature not supported on this platform")
				}

				if typescript != "" {
					if filepath.Ext(id) == ".ts" {
						transpileTypeScript()
					}

					// Watch all TypeScript files
					filepath.WalkDir(typescript, func(path string, dirEntry fs.DirEntry, err error) error {
						if (filepath.Ext(path) == ".ts") && !dirEntry.IsDir() {
							environment.Watch(path)
						}
						return nil
					})
				}
			}

			environment.ClearCache()
			_, err := environment.RequireID(startId)

			environment.Lock.Unlock()

			util.FailOnError(err)
		}

		environment.OnChanged("", nil)

		// Block forever
		select {}
	},
}

func transpileTypeScript() {
	log.Infof("transpiling TypeScript: %s", typescript)
	cmd := exec.Command("tsc", "--project", typescript)
	err := cmd.Run()
	util.FailOnError(err)
}
