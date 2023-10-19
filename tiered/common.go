package tiered

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/prudence/platform"
)

var log = commonlog.GetLogger("prudence.tiered")

func RegisterDefaultTypes() {
	platform.RegisterType("TieredCache", CreateTieredCacheBackend,
		"caches",
	)
}
