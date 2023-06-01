package js

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

func newExtensions(arguments map[string]string) []commonjs.Extension {
	var extensions []commonjs.Extension

	extensions = append(extensions, commonjs.Extension{
		Name:   "bind",
		Create: commonjs.CreateLateBindExtension,
	})

	extensions = append(extensions, commonjs.Extension{
		Name:   "prudence",
		Create: newPrudenceExtensionCreator(arguments),
	})

	extensions = append(extensions, commonjs.Extension{
		Name: "console",
		Create: func(context *commonjs.Context) goja.Value {
			return context.Environment.Runtime.ToValue(ConsoleAPI{})
		},
	})

	platform.OnAPIs(func(name string, api interface{}) bool {
		extensions = append(extensions, commonjs.Extension{
			Name: name,
			Create: func(context *commonjs.Context) goja.Value {
				return context.Environment.Runtime.ToValue(api)
			},
		})
		return true
	})

	return extensions
}

func newPrudenceExtensionCreator(arguments map[string]string) commonjs.CreateExtensionFunc {
	return func(context *commonjs.Context) goja.Value {
		prudence := context.Environment.Runtime.NewObject()

		// Copy API
		prudence_ := context.Environment.Runtime.ToValue(NewPrudenceAPI(context.Environment.URLContext, context, arguments)).ToObject(context.Environment.Runtime)
		for _, key := range prudence_.Keys() {
			prudence.Set(key, prudence_.Get(key))
		}

		// Globals
		prudence.Set("globals", Globals.NewDynamicObject(context.Environment.Runtime))

		// Type constructors
		platform.OnTypes(func(type_ string, create platform.CreateFunc) bool {
			prudence.Set(type_, newTypeConstructor(create, context))
			return true
		})

		return prudence
	}
}

func newTypeConstructor(create platform.CreateFunc, context *commonjs.Context) func(constructor goja.ConstructorCall) *goja.Object {
	runtime := context.Environment.Runtime
	// goja constructor signature
	return func(constructor goja.ConstructorCall) *goja.Object {
		var config ard.StringMap
		if len(constructor.Arguments) > 0 {
			config_ := constructor.Arguments[0].Export()
			var ok bool
			if config, ok = config_.(ard.StringMap); !ok {
				panic(runtime.NewGoError(fmt.Errorf("invalid \"config\" argument: %v", config_)))
			}
		} else {
			config = make(ard.StringMap)
		}

		if object, err := create(config, context); err == nil {
			return runtime.ToValue(object).ToObject(runtime)
		} else {
			panic(runtime.NewGoError(err))
		}
	}
}
