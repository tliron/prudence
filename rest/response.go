package rest

import (
	"bytes"
	"net/http"
	"time"

	"github.com/tliron/kutil/util"
)

//
// Response
//

type Response struct {
	Status      int
	Header      http.Header
	ContentType string
	CharSet     string
	Language    string

	Signature     string
	WeakSignature bool
	Timestamp     time.Time

	Direct http.ResponseWriter
	Buffer *bytes.Buffer
	Bypass bool
}

func NewResponse(responseWriter http.ResponseWriter) *Response {
	return &Response{
		Header: make(http.Header),
		Direct: responseWriter,
		Buffer: bytes.NewBuffer(nil),
	}
}

func (self *Response) Body() string {
	return util.BytesToString(self.Buffer.Bytes())
}

func (self *Response) flush() {
	if self.Bypass {
		return
	}

	status := self.Status
	if (status < 100) || (status > 999) {
		// Otherwise will panic in net/http.checkWriteHeaderCode
		status = http.StatusOK
	}

	header := self.Direct.Header()
	for name, values := range self.Header {
		for _, value := range values {
			header.Add(name, value)
		}
	}

	self.Direct.WriteHeader(status)
	self.Direct.Write(self.Buffer.Bytes())
}

func (self *Response) eTag() string {
	if self.Signature != "" {
		if self.WeakSignature {
			return "W/\"" + self.Signature + "\""
		} else {
			return "\"" + self.Signature + "\""
		}
	} else {
		return ""
	}
}

func (self *Response) setContentType() {
	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
	if self.ContentType != "" {
		if self.CharSet != "" {
			self.Header.Set("Content-Type", self.ContentType+";charset="+self.CharSet)
		} else {
			self.Header.Set("Content-Type", self.ContentType)
		}
	}
}

func (self *Response) setETag() {
	// ETag
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	if eTag := self.eTag(); eTag != "" {
		self.Header.Set("ETag", eTag)
	}
}

func (self *Response) setLastModified() {
	// Last-Modified
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	if !self.Timestamp.IsZero() {
		self.Header.Set("Last-Modified", self.Timestamp.Format(http.TimeFormat))
	}
}
