package rest

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"net/http"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
	"gocloud.dev/server/requestlog"
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
	NCSAPrefix  string
	Debug       bool
	Handler     HandleFunc

	server     *http.Server
	serverLock sync.Mutex
	started    sync.WaitGroup
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
	self.NCSAPrefix, _ = config_.Get("ncsa").String(true)
	self.Debug, _ = config_.Get("debug").Boolean(false)
	if handler := config_.Get("handler").Data; handler != nil {
		var err error
		if self.Handler, err = GetHandleFunc(handler, context); err != nil {
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

		var handler http.Handler = self

		// NCSA logging
		if logger, err := self.getNcsaLogger(); err == nil {
			if logger != nil {
				handler = requestlog.NewHandler(logger, handler)
			}
		}

		server := &http.Server{
			Addr:         self.Address,
			ReadTimeout:  time.Duration(time.Second * 5),
			WriteTimeout: time.Duration(time.Second * 5),
			Handler:      handler,
		}

		self.serverLock.Lock()
		self.server = server
		self.serverLock.Unlock()

		err = server.Serve(listener)

		if err == http.ErrServerClosed {
			err = nil
		}
	}

	return err
}

// Startable interface
func (self *Server) Stop() error {
	self.serverLock.Lock()
	defer self.serverLock.Unlock()

	if self.server != nil {
		log.Infof("stopping server: %s", self.Address)
		err := self.server.Shutdown(context.TODO())
		self.started.Wait()
		self.server = nil
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
		if err := context.Response.flush(); err != nil {
			log.Errorf("%s", err.Error())
		}
	}
}

var ncsaLoggers map[string]*requestlog.NCSALogger = make(map[string]*requestlog.NCSALogger)
var ncsaLoggersLock sync.Mutex

func (self *Server) getNcsaLogger() (*requestlog.NCSALogger, error) {
	var path string

	if platform.NCSAFilename != "" {
		switch platform.NCSAFilename {
		case "stdout", "stderr":
			path = platform.NCSAFilename
		default:
			path = self.NCSAPrefix + platform.NCSAFilename
		}
	}

	if path == "" {
		return nil, nil
	}

	var logger *requestlog.NCSALogger
	var ok bool

	ncsaLoggersLock.Lock()
	defer ncsaLoggersLock.Unlock()

	if logger, ok = ncsaLoggers[path]; !ok {
		var writer io.Writer

		switch path {
		case "stdout":
			writer = os.Stdout

		case "stderr":
			writer = os.Stderr

		default:
			if file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err == nil {
				util.OnExit(func() {
					file.Close()
				})
				writer = file
			} else {
				return nil, err
			}
		}

		logger = requestlog.NewNCSALogger(writer, func(err error) {
			log.Errorf("%s", err.Error())
		})

		ncsaLoggers[path] = logger
	}

	log.Infof("NCSA log for %s: %s", self.Address, path)

	return logger, nil
}
