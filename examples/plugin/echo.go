package plugin

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
	"github.com/tliron/prudence/rest"
)

func init() {
	platform.RegisterType("Echo", CreateEcho)
}

//
// Echo
//

type Echo struct {
	Message string
}

// platform.CreateFunc signature
func CreateEcho(config ard.StringMap, context *js.Context) (interface{}, error) {
	var self Echo
	config_ := ard.NewNode(config)
	self.Message, _ = config_.Get("message").String(true)
	return &self, nil
}

// rest.Handler interface
// rest.HandleFunc signature
func (self *Echo) Handle(context *rest.Context) bool {
	context.WriteString(self.Message + "\n")
	return true
}
