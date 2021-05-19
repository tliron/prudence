package js

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/platform"
)

func (self *PrudenceAPI) Create(config ard.StringMap) (interface{}, error) {
	return platform.Create(config, self.GetRelativeURL)
}

func (self *PrudenceAPI) Render(content string, renderer string) (string, error) {
	return platform.Render(content, renderer, self.GetRelativeURL)
}

func (self *PrudenceAPI) Start(startables interface{}) error {
	var list []ard.Value
	var ok bool
	if list, ok = startables.(ard.List); !ok {
		list = ard.List{startables}
	}

	var startables_ []platform.Startable

	for _, startable := range list {
		if startable_, ok := startable.(platform.Startable); ok {
			startables_ = append(startables_, startable_)
		} else {
			return fmt.Errorf("object not startable: %T", startable)
		}
	}

	return platform.Start(startables_)
}

func (self *PrudenceAPI) SetCache(cacheBackend platform.CacheBackend) {
	platform.SetCacheBackend(cacheBackend)
}
