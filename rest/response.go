package rest

import (
	"bytes"
	"net/http"
	"time"

	"github.com/tliron/go-ard"
)

//
// Response
//

type Response struct {
	Status      int
	Header      http.Header
	Cookies     []*http.Cookie
	ContentType string
	CharSet     string
	Language    string

	Signature     string
	WeakSignature bool
	Timestamp     time.Time

	Buffer *bytes.Buffer
	Bypass bool
	Direct http.ResponseWriter
}

func NewResponse(responseWriter http.ResponseWriter) *Response {
	return &Response{
		Header: make(http.Header),
		Direct: responseWriter,
		Buffer: bytes.NewBuffer(nil),
	}
}

func (self *Response) Reset() {
	self.Header = make(http.Header)
	self.Buffer.Reset()
}

func (self *Response) AddCookie(config ard.StringMap) error {
	if cookie, err := CreateCookie(config, nil); err == nil {
		self.Cookies = append(self.Cookies, cookie.(*http.Cookie))
		return nil
	} else {
		return err
	}
}

func (self *Response) flush() error {
	if self.Bypass {
		return nil
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

	for _, cookie := range self.Cookies {
		http.SetCookie(self.Direct, cookie)
	}

	self.Direct.WriteHeader(status)
	_, err := self.Direct.Write(self.Buffer.Bytes())
	return err
}

func (self *Response) eTag(fromHeader bool) string {
	if fromHeader {
		return self.Header.Get(HeaderETag)
	} else {
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
}

func (self *Response) lastModified(fromHeader bool) time.Time {
	if fromHeader {
		if lastModified, err := http.ParseTime(self.Header.Get(HeaderLastModified)); err == nil {
			return lastModified
		} else {
			return time.Time{}
		}
	} else {
		return self.Timestamp
	}
}

func (self *Response) setContentType() {
	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
	if self.ContentType != "" {
		if self.CharSet != "" {
			self.Header.Set(HeaderContentType, self.ContentType+";charset="+self.CharSet)
		} else {
			self.Header.Set(HeaderContentType, self.ContentType)
		}
	}
}

func (self *Response) setETag() {
	// ETag
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	if eTag := self.eTag(false); eTag != "" {
		self.Header.Set(HeaderETag, eTag)
	}
}

func (self *Response) setLastModified() {
	// Last-Modified
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	if !self.Timestamp.IsZero() {
		self.Header.Set(HeaderLastModified, self.Timestamp.Format(http.TimeFormat))
	}
}

//
// ResponseWriter
//

// Circumvents the built-in 404 response
// See: https://stackoverflow.com/a/47286697

type ResponseWriterWrapper struct {
	http.ResponseWriter
	context *Context
}

func NewResponseWriterWrapper(context *Context) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		ResponseWriter: context.Response.Direct,
		context:        context,
	}
}

// http.ResponseWriter interface
func (self *ResponseWriterWrapper) WriteHeader(status int) {
	// Store response in context
	self.context.Response.Status = status

	// Don't write the 404 header
	if status != http.StatusNotFound {
		self.ResponseWriter.WriteHeader(status)
	}
}

// http.ResponseWriter interface
func (self *ResponseWriterWrapper) Write(p []byte) (int, error) {
	if self.context.Response.Status != http.StatusNotFound {
		return self.ResponseWriter.Write(p)
	} else {
		// Don't write the 404 response but pretend that we did
		return len(p), nil
	}
}
