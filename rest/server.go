package rest

import (
	contextpkg "context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
	"gocloud.dev/server/requestlog"
)

const (
	// See: https://ieftimov.com/posts/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/
	DEFAULT_HANDLER_TIMEOUT     = 5 * time.Second  // within the write timeout
	DEFAULT_READ_HEADER_TIMEOUT = 3 * time.Second  // from end of TLS handshake until start of client sending request
	DEFAULT_READ_TIMEOUT        = 10 * time.Second // from beginning of connection until end of client sending request
	DEFAULT_WRITE_TIMEOUT       = 10 * time.Second // from start of client sending request until end of server sending response
	DEFAULT_IDLE_TIMEOUT        = 30 * time.Second // from end of server sending response until next request (keepalive only)
)

//
// Server
//

type Server struct {
	Name                string
	Protocol            string
	Address             string
	Port                uint64
	TLS                 bool
	Certificate         string
	Key                 string
	GenerateCertificate bool
	NCSALogFileSuffix   string
	Debug               bool
	HandlerTimeout      time.Duration
	ReadHeaderTimeout   time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	Handler             HandleFunc

	server     *http.Server
	serverLock sync.Mutex
	started    sync.WaitGroup
	log        commonlog.Logger
}

func NewServer(name string) *Server {
	if name == "" {
		name = "Prudence"
	}

	return &Server{
		Name:              name,
		Port:              8080,
		log:               log,
		HandlerTimeout:    DEFAULT_HANDLER_TIMEOUT,
		ReadHeaderTimeout: DEFAULT_READ_HEADER_TIMEOUT,
		ReadTimeout:       DEFAULT_READ_TIMEOUT,
		WriteTimeout:      DEFAULT_WRITE_TIMEOUT,
		IdleTimeout:       DEFAULT_IDLE_TIMEOUT,
	}
}

// ([platform.CreateFunc] signature)
func CreateServer(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	name, _ := config_.Get("name").String()

	self := NewServer(name)

	address, _ := config_.Get("address").String()
	if address, addressZone, err := util.ToReachableIPAddress(address); err == nil {
		if addressZone != "" {
			address += "%" + addressZone
		}

		self.Address = address
	} else {
		return nil, err
	}

	if port, ok := config_.Get("port").UnsignedInteger(); ok {
		self.Port = port
	}

	protocol, _ := config_.Get("protocol").String()
	switch strings.ToLower(protocol) {
	case "":
		self.Protocol = "dual"

	case "dual", "ipv6", "ipv4":
		self.Protocol = protocol

	default:
		return nil, fmt.Errorf("\"ip\" must be \"dual\", \"ipv6\" or \"ipv4\": %s", protocol)
	}

	tls := config_.Get("tls")
	self.Certificate, _ = tls.Get("certificate").String()
	self.Key, _ = tls.Get("key").String()
	self.GenerateCertificate, _ = tls.Get("generate").Boolean()
	if self.GenerateCertificate || (self.Certificate != "") || (self.Key != "") {
		self.TLS = true
	}

	self.NCSALogFileSuffix, _ = config_.Get("ncsaLogFileSuffix").String()
	self.Debug, _ = config_.Get("debug").Boolean()

	if timeout, ok := config_.Get("handlerTimeout").Float(); ok {
		self.HandlerTimeout = time.Duration(timeout * float64(time.Second))
	}

	if timeout, ok := config_.Get("readHeaderTimeout").Float(); ok {
		self.ReadHeaderTimeout = time.Duration(timeout * float64(time.Second))
	}

	if timeout, ok := config_.Get("readTimeout").Float(); ok {
		self.ReadTimeout = time.Duration(timeout * float64(time.Second))
	}

	if timeout, ok := config_.Get("writeTimeout").Float(); ok {
		self.WriteTimeout = time.Duration(timeout * float64(time.Second))
	}

	if timeout, ok := config_.Get("idleTimeout").Float(); ok {
		self.IdleTimeout = time.Duration(timeout * float64(time.Second))
	}

	if handler := config_.Get("handler").Value; handler != nil {
		var err error
		if self.Handler, err = GetHandleFunc(handler, jsContext); err != nil {
			return nil, err
		}
	}

	return self, nil
}

// ([platform.Startable] interface)
func (self *Server) Start() error {
	self.log = commonlog.NewKeyValueLogger(log,
		"_scope", "server",
		"name", self.Name,
		"address", self.Address,
		"port", self.Port,
		"protocol", self.Protocol,
		"secure", self.TLS,
	)

	self.started.Add(1)
	defer self.started.Done()

	self.log.Info("starting")

	if listener, err := self.newListener(); err == nil {
		defer listener.Close()

		var handler http.Handler = self

		if logger := self.newNcsaLogger(); logger != nil {
			handler = requestlog.NewHandler(logger, handler)
		}

		handler = http.TimeoutHandler(handler, self.HandlerTimeout, "")

		server := &http.Server{
			Addr:              self.Address,
			ReadHeaderTimeout: self.ReadHeaderTimeout,
			ReadTimeout:       self.ReadTimeout,
			WriteTimeout:      self.WriteTimeout,
			IdleTimeout:       self.IdleTimeout,
			Handler:           handler,
		}

		if self.log.AllowLevel(commonlog.Debug) {
			server.ConnState = func(conn net.Conn, state http.ConnState) {
				self.log.Debug(state.String(), "_scope", "connection")
			}
		}

		self.serverLock.Lock()
		self.server = server
		self.serverLock.Unlock()

		if err := server.Serve(listener); (err == nil) || (err == http.ErrServerClosed) {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

// ([platform.Startable] interface)
func (self *Server) Stop(stopContext contextpkg.Context) error {
	self.serverLock.Lock()
	defer self.serverLock.Unlock()

	if self.server != nil {
		self.log.Info("stopping")
		err := self.server.Shutdown(stopContext)
		self.started.Wait()
		self.server = nil
		self.log.Info("stopped")
		return err
	} else {
		return nil
	}
}

// ([http.Handler] interface)
func (self *Server) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if self.Handler == nil {
		return
	}

	restContext := NewContext(responseWriter, request, self.log)

	defer func() {
		if r := recover(); r != nil {
			if r == EndRequest {
				restContext.Log.Debug("end")
				if err := restContext.Response.flush(); err != nil {
					self.log.Error(err.Error())
				}
			} else {
				panic(r)
			}
		}
	}()

	restContext.Debug = self.Debug

	if self.Name != "" {
		restContext.Response.StaticHeader.Set(HeaderServer, self.Name)
	}

	if _, err := self.Handler(restContext); err != nil {
		restContext.Response.Reset()
		if restContext.Debug {
			restContext.Write(err.Error())
			restContext.Write("\n")
		}
		restContext.InternalServerError(err)
	}

	if err := restContext.Response.flush(); err != nil {
		self.log.Error(err.Error())
	}
}

func (self *Server) AddressPort() string {
	return util.JoinIPAddressPort(self.Address, int(self.Port))
}

func (self *Server) newListener() (net.Listener, error) {
	protocol := "tcp"
	switch strings.ToLower(self.Protocol) {
	case "ipv6":
		protocol = "tcp6"

	case "ipv4":
		protocol = "tcp4"
	}

	if tlsConfig, err := self.newTlsConfig(); err == nil {
		if listener, err := net.Listen(protocol, self.AddressPort()); err == nil {
			if tlsConfig != nil {
				return tls.NewListener(listener, tlsConfig), nil
			} else {
				return listener, nil
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Server) newTlsConfig() (*tls.Config, error) {
	if !self.TLS {
		return nil, nil
	}

	log := commonlog.NewKeyValueLogger(self.log, "_scope", "tls")

	var tlsConfig *tls.Config
	var err error

	if self.GenerateCertificate {
		log.Info("generating certificate and key")
		if tlsConfig, err = util.CreateSelfSignedTLSConfig("Prudence", self.Address, 0, 0); err == nil {
			if len(tlsConfig.Certificates) == 0 {
				panic("no TLS certificates")
			}

			write := func(path string, f func(file *os.File) error) error {
				if file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600); err == nil {
					defer commonlog.CallAndLogWarning(file.Close, "Server.newTlsConfig", log)
					return f(file)
				} else {
					return err
				}
			}

			if self.Certificate != "" {
				log.Infof("writing certificate to: %s" + self.Certificate)
				if err := write(self.Certificate, func(file *os.File) error {
					return util.WriteTLSCertificatePEM(file, &tlsConfig.Certificates[0])
				}); err != nil {
					return nil, err
				}
			}

			if self.Key != "" {
				log.Infof("writing key to: %s", self.Key)
				if err := write(self.Key, func(file *os.File) error {
					return util.WriteTLSRSAKeyPEM(file, &tlsConfig.Certificates[0])
				}); err != nil {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	} else {
		log.Info("using provided certificate and key")
		if tlsConfig, err = util.CreateTLSConfig(util.StringToBytes(self.Certificate), (util.StringToBytes(self.Key))); err != nil {
			return nil, err
		}
	}

	// This is *not* set with "h2" when calling Server.Serve!
	tlsConfig.NextProtos = []string{"h2", "http/1.1"}

	return tlsConfig, nil
}

var ncsaLoggers map[string]*requestlog.NCSALogger = make(map[string]*requestlog.NCSALogger)
var ncsaLoggersLock sync.Mutex

func (self *Server) newNcsaLogger() *requestlog.NCSALogger {
	path := platform.NCSAFilename

	if path == "" {
		return nil
	} else {
		// Don't use suffix for "/dev/*" paths
		if !strings.HasPrefix(path, "/dev/") {
			ext := filepath.Ext(path)
			path = path[:len(path)-len(ext)] + self.NCSALogFileSuffix + ext
		}
	}

	log := commonlog.NewKeyValueLogger(self.log, "_scope", "ncsa")

	var logger *requestlog.NCSALogger
	var ok bool

	ncsaLoggersLock.Lock()
	defer ncsaLoggersLock.Unlock()

	if logger, ok = ncsaLoggers[path]; !ok {
		var writer io.Writer

		if file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err == nil {
			util.OnExitError(file.Close)
			writer = file
		} else {
			log.Error(err.Error())
			return nil
		}

		logger = requestlog.NewNCSALogger(writer, func(err error) {
			log.Error(err.Error())
		})

		ncsaLoggers[path] = logger
	}

	log.Info(path)

	return logger
}
