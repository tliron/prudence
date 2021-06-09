package rest

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tliron/kutil/util"
)

//
// Request
//

type Request struct {
	Host    string
	Port    int
	Header  http.Header
	Method  string
	Query   url.Values
	Cookies []*http.Cookie
	Body    string

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
		Header:  request.Header,
		Method:  request.Method,
		Query:   request.URL.Query(),
		Cookies: request.Cookies(),
		Direct:  request,
	}

	if request.Body != nil {
		defer request.Body.Close()
		if bytes, err := io.ReadAll(request.Body); err == nil {
			self.Body = util.BytesToString(bytes)
		}
	}

	return &self
}
