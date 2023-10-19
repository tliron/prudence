package distributed

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/prudence/platform"
)

var log = commonlog.GetLogger("prudence.distributed")

func RegisterDefaultTypes() {
	platform.RegisterType("DistributedCache", CreateDistributedCacheBackend,
		"local",
		"kubernetes",
	)
}
