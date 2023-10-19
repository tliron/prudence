package js

import (
	"errors"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/distributed"
	"github.com/tliron/prudence/local"
	"github.com/tliron/prudence/memory"
	"github.com/tliron/prudence/platform"
	"github.com/tliron/prudence/rest"
	"github.com/tliron/prudence/tiered"
)

const DEFAULT_START_TIMEOUT_SECONDS = 10.0

func init() {
	rest.RegisterDefaultTypes()
	distributed.RegisterDefaultTypes()
	local.RegisterDefaultTypes()
	memory.RegisterDefaultTypes()
	tiered.RegisterDefaultTypes()
}

// ([commonjs.CreateExtensionFunc] signature)
func CreatePrudenceExtension(jsContext *commonjs.Context) any {
	prudence_ := NewPrudenceAPI(jsContext)
	prudence := commonjs.NewObject(jsContext.Environment.Runtime, prudence_)

	// Type constructors
	for typeName, type_ := range platform.Types {
		prudence.Set(typeName, newTypeConstructor(jsContext, type_))
	}

	return prudence
}

//
// PrudenceAPI
//

type PrudenceAPI struct {
	DataContentTypes      []string
	NotFound              rest.HandleFunc
	RedirectTrailingSlash rest.HandleFunc

	jsContext *commonjs.Context
}

func NewPrudenceAPI(jsContext *commonjs.Context) *PrudenceAPI {
	return &PrudenceAPI{
		DataContentTypes:      rest.DataContentTypes,
		NotFound:              rest.HandleNotFound,
		RedirectTrailingSlash: rest.HandleRedirectTrailingSlash,
		jsContext:             jsContext,
	}
}

func (self *PrudenceAPI) Start(startables any, stopTimeoutSeconds float64) error {
	var startables_ []platform.Startable

	addStartable := func(object any) bool {
		added := false

		if hasStartables, ok := object.(platform.HasStartables); ok {
			startables_ = append(startables_, hasStartables.GetStartables()...)
			added = true
		}

		if startable, ok := object.(platform.Startable); ok {
			startables_ = append(startables_, startable)
			added = true
		}

		return added
	}

	addStartable(platform.GetCacheBackend())
	addStartable(platform.GetScheduler())

	var list []any
	if list_, ok := startables.([]any); ok {
		list = list_
	} else {
		list = []any{startables}
	}

	for _, startable := range list {
		if !addStartable(startable) {
			return fmt.Errorf("object not startable: %T", startable)
		}
	}

	if stopTimeoutSeconds == 0.0 {
		stopTimeoutSeconds = DEFAULT_START_TIMEOUT_SECONDS
	}

	return platform.Start(startables_, time.Duration(stopTimeoutSeconds*float64(time.Second)))
}

func (self *PrudenceAPI) SetCache(cacheBackend platform.CacheBackend) {
	platform.SetCacheBackend(cacheBackend)
}

func (self *PrudenceAPI) InvalidateCache(key string) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheBackend.DeleteRepresentation(platform.CacheKey(key))
	}
}

func (self *PrudenceAPI) InvalidateCacheGroup(group string) {
	if cacheBackend := platform.GetCacheBackend(); cacheBackend != nil {
		cacheBackend.DeleteGroup(platform.CacheKey(group))
	}
}

func (self *PrudenceAPI) SetScheduler(scheduler platform.Scheduler) {
	platform.SetScheduler(scheduler)
}

func (self *PrudenceAPI) Schedule(cronPattern string, job func()) error {
	if scheduler := platform.GetScheduler(); scheduler != nil {
		return scheduler.Schedule(cronPattern, job)
	} else {
		return errors.New("no scheduler")
	}
}

// Utils

func newTypeConstructor(jsContext *commonjs.Context, type_ *platform.Type) commonjs.JavaScriptConstructorFunc {
	return commonjs.NewConstructor(jsContext.Environment.Runtime, func(constructor goja.ConstructorCall) (any, error) {
		var config ard.StringMap

		switch length := len(constructor.Arguments); length {
		case 0:
			config = make(ard.StringMap)

		case 1:
			config_ := constructor.Arguments[0].Export()
			var ok bool
			if config, ok = config_.(ard.StringMap); !ok {
				return nil, fmt.Errorf("%s constructor \"config\" argument is not an object: %T", type_.Name, config_)
			}

		default:
			return nil, fmt.Errorf("%s constructor has more than one argument: %d", type_.Name, length)
		}

		return type_.Create(jsContext, config)
	})
}
