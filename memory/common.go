package memory

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/prudence/platform"
)

var log = commonlog.GetLogger("prudence.memory")

func RegisterDefaultTypes() {
	platform.RegisterType("MapCache", CreateMapCacheBackend,
		"pruneFrequency",
	)

	platform.RegisterType("MemoryCache", CreateMemoryCacheBackend,
		"maxSize",
		"averageSize",
		"pruneFrequency",
	)
}
