package js

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/js"
)

var log = commonlog.GetLogger("prudence.js")

var Globals = js.NewThreadSafeObject()
