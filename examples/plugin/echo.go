package plugin

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/platform"
	"github.com/tliron/prudence/rest"
)

func init() {
	platform.RegisterType("myplugin.echo", CreateEcho)
}

//
// Echo
//

type Echo struct {
	Message string
}

// CreateFunc signature
func CreateEcho(config ard.StringMap, getRelativeURL platform.GetRelativeURL) (interface{}, error) {
	var self Echo
	config_ := ard.NewNode(config)
	self.Message, _ = config_.Get("message").String(true)
	return &self, nil
}

// Handler interface
// HandleFunc signature
func (self *Echo) Handle(context *rest.Context) bool {
	context.WriteString(self.Message + "\n")
	return true
}
