package rest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/fasthttp/http2"
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
func CreateServer(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
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
			cert, priv, err := GenerateTestCertificate(self.Address)
			if err != nil {
				return err
			}

			err = self.server.AppendCertEmbed(cert, priv)
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

func GenerateTestCertificate(host string) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"fasthttp test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		DNSNames:              []string{host},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(
		rand.Reader, cert, cert, &priv.PublicKey, priv,
	)

	p := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	b := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		},
	)

	return b, p, err
}
