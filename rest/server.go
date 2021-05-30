package rest

import (
	"net"

	"github.com/dop251/goja"
	"github.com/fasthttp/http2"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

func init() {
	platform.RegisterType("Server", CreateServer)
}

//
// Server
//

type Server struct {
	Name     string
	Address  string
	Protocol string
	Secure   bool
	Debug    bool
	Handler  HandleFunc

	server fasthttp.Server
}

func NewServer(name string) *Server {
	if name == "" {
		name = "Prudence"
	}
	return &Server{
		Name: name,
	}
}

// CreateFunc signature
func CreateServer(config ard.StringMap, resolve js.ResolveFunc, runtime *goja.Runtime) (interface{}, error) {
	var self Server

	config_ := ard.NewNode(config)
	self.Address, _ = config_.Get("address").String(false)
	self.Protocol, _ = config_.Get("protocol").String(false)
	if self.Protocol == "" {
		self.Protocol = "http"
	}
	self.Secure, _ = config_.Get("secure").Boolean(false)
	self.Debug, _ = config_.Get("debug").Boolean(false)
	if handler := config_.Get("handler").Data; handler != nil {
		var err error
		if self.Handler, err = GetHandleFunc(handler); err != nil {
			return nil, err
		}
	}
	self.Name, _ = config_.Get("name").String(false)
	if self.Name == "" {
		self.Name = "Prudence"
	}

	return &self, nil
}

func (self *Server) Listen() (net.Listener, error) {
	return net.Listen("tcp4", self.Address)
}

// Startable interface
func (self *Server) Start() error {
	log.Infof("starting server: %s", self.Address)

	if listener, err := self.Listen(); err == nil {
		self.server = fasthttp.Server{
			Handler:                       self.Handle,
			Name:                          self.Name,
			LogAllErrors:                  true,
			Logger:                        Logger{logging.GetLogger("prudence.server")},
			DisableHeaderNamesNormalizing: true,
			NoDefaultContentType:          true,
		}

		if self.Secure {
			certificate, privateKey, err := util.CreateSelfSignedX509("Prudence", self.Address)
			if err != nil {
				return err
			}

			err = self.server.AppendCertEmbed(certificate, privateKey)
			if err != nil {
				return err
			}

			if self.Protocol == "http2" {
				// TODO: BROKEN
				http2.ConfigureServer(&self.server)
			}

			if err := self.server.ServeTLS(listener, "", ""); err == nil {
				log.Infof("server stopped: %s", self.Address)
				return nil
			} else {
				return err
			}
		} else {
			if err := self.server.Serve(listener); err == nil {
				log.Infof("server stopped: %s", self.Address)
				return nil
			} else {
				return err
			}
		}
	} else {
		return err
	}
}

// Startable interface
func (self *Server) Stop() error {
	log.Infof("stopping server: %s", self.Address)
	return self.server.Shutdown()
}

// fasthttp.RequestHandler signature
func (self *Server) Handle(context *fasthttp.RequestCtx) {
	if self.Handler != nil {
		context_ := NewContext(context)
		context_.Debug = self.Debug
		self.Handler(context_)
	}
}
