package rest

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"net/http"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Server", CreateServer)
}

//
// Server
//

type Server struct {
	Name        string
	Address     string
	Secure      bool
	Certificate string
	Key         string
	Debug       bool
	Handler     HandleFunc

	server  *http.Server
	started sync.WaitGroup
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
func CreateServer(config ard.StringMap, context *js.Context) (interface{}, error) {
	var self Server

	config_ := ard.NewNode(config)
	self.Name, _ = config_.Get("name").String(false)
	if self.Name == "" {
		self.Name = "Prudence"
	}
	self.Address, _ = config_.Get("address").String(false)
	secure := config_.Get("secure")
	if secure.Data != nil {
		self.Secure = true
	}
	self.Certificate, _ = secure.Get("certificate").String(true)
	self.Key, _ = secure.Get("key").String(true)
	self.Debug, _ = config_.Get("debug").Boolean(false)
	if handler := config_.Get("handler").Data; handler != nil {
		var err error
		if self.Handler, err = GetHandleFunc(handler); err != nil {
			return nil, err
		}
	}

	return &self, nil
}

func (self *Server) newListener(secure bool) (net.Listener, error) {
	if listener, err := net.Listen("tcp", self.Address); err == nil {
		if secure {
			var tlsConfig *tls.Config
			var err error
			if (self.Certificate != "") || (self.Key != "") {
				if tlsConfig, err = util.CreateTLSConfig(util.StringToBytes(self.Certificate), (util.StringToBytes(self.Key))); err != nil {
					return nil, err
				}
			} else if tlsConfig, err = util.CreateSelfSignedTLSConfig("Prudence", self.Address); err != nil {
				return nil, err
			}

			// This is *not* set with "h2" when calling Server.Serve!
			tlsConfig.NextProtos = []string{"h2", "http/1.1"}

			return tls.NewListener(listener, tlsConfig), nil
		} else {
			return listener, nil
		}
	} else {
		return nil, err
	}
}

// Startable interface
func (self *Server) Start() error {
	self.started.Add(1)
	defer self.started.Done()

	if self.Secure {
		log.Infof("starting secure server: %s", self.Address)
	} else {
		log.Infof("starting server: %s", self.Address)
	}

	var err error
	var listener net.Listener
	if listener, err = self.newListener(self.Secure); err == nil {
		defer listener.Close()

		self.server = &http.Server{
			Addr:         self.Address,
			ReadTimeout:  time.Duration(time.Second * 5),
			WriteTimeout: time.Duration(time.Second * 5),
			Handler:      self,
		}

		err = self.server.Serve(listener)

		if err == http.ErrServerClosed {
			err = nil
			self.server = nil
		}
	}

	return err
}

// Startable interface
func (self *Server) Stop() error {
	if self.server != nil {
		log.Infof("stopping server: %s", self.Address)
		err := self.server.Shutdown(context.TODO())
		self.started.Wait()
		log.Infof("stopped server: %s", self.Address)
		return err
	} else {
		return nil
	}
}

// http.Handler interface
func (self *Server) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if self.Handler != nil {
		context := NewContext(responseWriter, request)
		if self.Name != "" {
			context.Response.Header.Set(HeaderServer, self.Name)
		}
		context.Debug = self.Debug
		self.Handler(context)
		context.Response.flush()
	}
}
