package js

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/commonlog"
)

var log = commonlog.GetLogger("prudence.js")

var Globals = commonjs.NewThreadSafeObject()
