package rest

import (
	"io"
	"net/http"
	"net/url"

	"github.com/tliron/kutil/util"
)

//
// Request
//

type Request struct {
	URL    *url.URL
	Header http.Header
	Method string
	Query  url.Values
	Body   string

	Direct *http.Request
}

func NewRequest(request *http.Request) *Request {
	self := Request{
		URL:    request.URL,
		Header: request.Header,
		Method: request.Method,
		Query:  request.URL.Query(),
		Direct: request,
	}

	if request.Body != nil {
		defer request.Body.Close()
		if bytes, err := io.ReadAll(request.Body); err == nil {
			self.Body = util.BytesToString(bytes)
		}
	}

	return &self
}
