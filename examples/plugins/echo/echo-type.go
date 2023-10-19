package echo

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
	"github.com/tliron/prudence/rest"
)

func init() {
	platform.RegisterType("Echo", CreateEcho,
		// Allowed config keys
		"message",
	)
}

//
// Echo
//

type Echo struct {
	Message string
}

// ([platform.CreateFunc] signature)
func CreateEcho(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()
	message, _ := config_.Get("message").String()

	return &Echo{
		Message: message,
	}, nil
}

// ([rest.Handler] interface, [rest.HandleFunc] signature)
func (self *Echo) Handle(restContext *rest.Context) (bool, error) {
	return true, restContext.Write(self.Message + "\n")
}
