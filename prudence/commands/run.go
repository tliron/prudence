package commands

import (
	contextpkg "context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/exturl"
	"github.com/tliron/kutil/util"
	"github.com/tliron/kutil/version"
	"github.com/tliron/prudence/js"
	"github.com/tliron/prudence/platform"
)

var paths []string
var useWorkingDir bool
var typescript string
var arguments map[string]string
var watch bool

func init() {
	rootCommand.AddCommand(runCommand)
	runCommand.Flags().StringArrayVarP(&paths, "path", "p", nil, "library path (appended after PRUDENCE_PATH environment variable)")
	runCommand.Flags().BoolVarP(&useWorkingDir, "use-working-dir", "d", true, "whether to include the current working dir in the library path")
	runCommand.Flags().StringVarP(&typescript, "typescript", "t", "", "TypeScript project path (must have a \"tsconfig.json\" file)")
	runCommand.Flags().StringToStringVarP(&arguments, "argument", "a", make(map[string]string), "arguments (format is name=value)")
	runCommand.Flags().BoolVarP(&watch, "watch", "w", true, "whether to watch dependent files and restart if they are changed")
	runCommand.Flags().StringVarP(&platform.NCSAFilename, "ncsa", "n", "", "NCSA log filename")
}

var runCommand = &cobra.Command{
	Use:   "run [Script PATH or URL]",
	Short: "Run",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startId := args[0]

		util.OnExit(platform.Stop)

		urlContext := exturl.NewContext()
		util.OnExitError(urlContext.Release)

		var bases []exturl.URL
		var basePaths []exturl.URL

		if useWorkingDir {
			workingDirFileUrl, err := urlContext.NewWorkingDirFileURL()
			util.FailOnError(err)
			log.Infof("work dir: %s", workingDirFileUrl.String())
			bases = []exturl.URL{workingDirFileUrl}
			basePaths = []exturl.URL{workingDirFileUrl}
		}

		addBasePaths := func(paths []string) {
			for _, path := range paths {
				if !strings.HasSuffix(path, "/") {
					path += "/"
				}
				pathUrl, err := urlContext.NewValidAnyOrFileURL(contextpkg.TODO(), path, bases)
				util.FailOnError(err)
				log.Infof("library path: %s", pathUrl.String())
				basePaths = append(basePaths, pathUrl)
			}
		}

		addBasePaths(filepath.SplitList(os.Getenv("PRUDENCE_PATH")))
		addBasePaths(paths)

		environment := js.NewEnvironment(arguments, urlContext, basePaths...)
		util.OnExitError(environment.Release)

		log.Noticef("Prudence version: %s", version.GitVersion)

		if typescript != "" {
			transpileTypeScript()
		}

		environment.OnFileModified = func(id string, module *commonjs.Module) {
			if module != nil {
				log.Infof("module changed: %s", module.Id)
			} else if id != "" {
				log.Infof("file changed: %s", id)
			}

			environment.Lock.Lock()

			if watch {
				if err := environment.StartWatcher(); err != nil {
					log.Error(err.Error())
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
			_, err := environment.Require(startId, false, nil)

			environment.Lock.Unlock()

			util.FailOnError(err)
		}

		environment.OnFileModified("", nil)

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
