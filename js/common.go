package js

import (
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("prudence.js")

var Globals = js.NewThreadSafeObject()
