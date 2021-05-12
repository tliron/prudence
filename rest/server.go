package rest

import (
	"net"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/js/common"
	"github.com/valyala/fasthttp"
)

func init() {
	Register("server", CreateServer)
}

//
// Server
//

type Server struct {
	Name    string
	Address string
	Handler HandleFunc
}

func NewServer(address string, handler HandleFunc) *Server {
	return &Server{
		Name:    "Prudence",
		Address: address,
		Handler: handler,
	}
}

// CreateFunc signature
func CreateServer(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	var self Server

	config_ := ard.NewNode(config)
	self.Address, _ = config_.Get("address").String(false)
	handler := config_.Get("handler").Data
	self.Handler, _ = GetHandleFunc(handler)
	self.Name, _ = config_.Get("name").String(false)
	if self.Name == "" {
		self.Name = "Prudence"
	}

	return &self, nil
}

func (self *Server) Listen() (net.Listener, error) {
	return net.Listen("tcp4", self.Address)
}

func (self *Server) Start() error {
	log.Infof("starting server: %s", self.Address)

	if listener, err := self.Listen(); err == nil {
		server := fasthttp.Server{
			Handler:                       self.Handle,
			Name:                          self.Name,
			LogAllErrors:                  true,
			Logger:                        Logger{},
			DisableHeaderNamesNormalizing: true,
			NoDefaultContentType:          true,
		}
		return server.Serve(listener)
	} else {
		return err
	}
}

// fasthttp.RequestHandler signature
func (self *Server) Handle(context *fasthttp.RequestCtx) {
	if self.Handler != nil {
		self.Handler(NewContext(context))
	}
}
