package rest

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

//
// Request
//

type Request struct {
	Host    string
	Port    int
	Path    string
	Header  http.Header
	Method  string
	Query   url.Values
	Cookies []*http.Cookie

	Direct *http.Request
}

func NewRequest(request *http.Request) *Request {
	host := request.Host
	var port int
	if colon := strings.IndexRune(host, ':'); colon != -1 {
		port, _ = strconv.Atoi(host[colon+1:])
		host = host[:colon]
	}

	self := Request{
		Host:    host,
		Port:    port,
		Path:    strings.TrimPrefix(request.URL.Path, "/"),
		Header:  request.Header.Clone(),
		Method:  request.Method,
		Query:   request.URL.Query(), // will be a new map
		Cookies: request.Cookies(),   // will be a new array
		Direct:  request,
	}

	return &self
}

func (self *Request) Clone() *Request {
	query := make(url.Values)
	for name, values := range self.Query {
		query[name] = append(values[:0:0], values...)
	}

	return &Request{
		Host:    self.Host,
		Port:    self.Port,
		Path:    self.Path,
		Header:  self.Header.Clone(),
		Method:  self.Method,
		Query:   CloneURLValues(self.Query),
		Cookies: CloneCookies(self.Cookies),
		Direct:  self.Direct,
	}
}

func (self *Request) Body() ([]byte, error) {
	if self.Direct.Body != nil {
		defer commonlog.CallAndLogWarning(self.Direct.Body.Close, "Request.Body", log)
		// TODO: limit size of requests
		return io.ReadAll(self.Direct.Body)
	} else {
		return nil, nil
	}
}

func (self *Request) BodyAsString() (string, error) {
	if body, err := self.Body(); err == nil {
		return util.BytesToString(body), nil
	} else {
		return "", err
	}
}

func (self *Request) GetCookie(name string) *http.Cookie {
	for _, cookie := range self.Cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
