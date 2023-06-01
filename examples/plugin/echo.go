package plugin

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
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
func CreateEcho(config ard.StringMap, context *commonjs.Context) (interface{}, error) {
	var self Echo
	config_ := ard.NewNode(config)
	self.Message, _ = config_.Get("message").String()
	return &self, nil
}

// rest.Handler interface
// rest.HandleFunc signature
func (self *Echo) Handle(context *rest.Context) bool {
	context.WriteString(self.Message + "\n")
	return true
}
