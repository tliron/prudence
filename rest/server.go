package rest

import (
	"net"

	"github.com/fasthttp/http2"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

func init() {
	platform.RegisterType("server", CreateServer)
}

//
// Server
//

type Server struct {
	Name     string
	Address  string
	Protocol string
	Secure   bool
	Handler  HandleFunc

	server fasthttp.Server
}

func NewServer(address string, handler HandleFunc) *Server {
	return &Server{
		Name:    "Prudence",
		Address: address,
		Handler: handler,
	}
}

// CreateFunc signature
func CreateServer(config ard.StringMap, getRelativeURL platform.GetRelativeURL) (interface{}, error) {
	var self Server

	config_ := ard.NewNode(config)
	self.Address, _ = config_.Get("address").String(false)
	self.Protocol, _ = config_.Get("protocol").String(false)
	if self.Protocol == "" {
		self.Protocol = "http"
	}
	self.Secure, _ = config_.Get("tls").Boolean(false)
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

// Startable interface
func (self *Server) Start() error {
	log.Infof("starting server: %s", self.Address)

	if listener, err := self.Listen(); err == nil {
		self.server = fasthttp.Server{
			Handler:                       self.Handle,
			Name:                          self.Name,
			LogAllErrors:                  true,
			Logger:                        Logger{},
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
				// STILL BROKEN
				http2.ConfigureServer(&self.server)
			}

			return self.server.ServeTLS(listener, "", "")
		} else {
			return self.server.Serve(listener)
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
		self.Handler(NewContext(context))
	}
}
