package local

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/prudence/platform"
)

var log = commonlog.GetLogger("prudence.local")

func RegisterDefaultTypes() {
	platform.RegisterType("LocalScheduler", CreateLocalScheduler)
}
