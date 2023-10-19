package rest

import (
	"bytes"
	"net/http"
	"time"

	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

//
// Response
//

type Response struct {
	Status       int
	Header       http.Header
	StaticHeader http.Header
	Cookies      []*http.Cookie
	ContentType  string
	CharSet      string
	Language     string

	Signature     string
	WeakSignature bool
	Timestamp     time.Time

	Buffer *bytes.Buffer
	Bypass bool
	Direct http.ResponseWriter
}

func NewResponse(responseWriter http.ResponseWriter) *Response {
	return &Response{
		Header:       make(http.Header),
		StaticHeader: make(http.Header),
		CharSet:      "utf-8",
		Direct:       responseWriter,
		Buffer:       bytes.NewBuffer(nil),
	}
}

func (self *Response) Clone() *Response {
	return &Response{
		Status:        self.Status,
		Header:        self.Header.Clone(),
		StaticHeader:  self.StaticHeader.Clone(),
		Cookies:       CloneCookies(self.Cookies),
		ContentType:   self.ContentType,
		CharSet:       self.CharSet,
		Language:      self.Language,
		Signature:     self.Signature,
		WeakSignature: self.WeakSignature,
		Timestamp:     self.Timestamp,
		Buffer:        self.Buffer,
		Bypass:        self.Bypass,
		Direct:        self.Direct,
	}
}

func (self *Response) Reset() {
	self.Header = make(http.Header)
	self.Buffer.Reset()
}

func (self *Response) AddCookie(config ard.StringMap) error {
	if cookie, err := platform.Create(nil, "Cookie", config); err == nil {
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
	for name, values := range self.StaticHeader {
		for _, value := range values {
			header.Add(name, value)
		}
	}
	for name, values := range self.Header {
		for _, value := range values {
			header.Add(name, value)
		}
	}

	for _, cookie := range self.Cookies {
		http.SetCookie(self.Direct, cookie)
	}

	if self.Buffer.Len() == 0 {
		/*if status == http.StatusOK {
			status = http.StatusNoContent // 204
		}*/
		self.Direct.WriteHeader(status)
		return nil
	} else {
		self.Direct.WriteHeader(status)
		_, err := self.Direct.Write(self.Buffer.Bytes())
		return err
	}
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
		lastModified, _ := GetTimeHeader(HeaderLastModified, self.Header)
		return lastModified
	} else {
		return self.Timestamp
	}
}

func (self *Response) setContentType() {
	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
	if self.ContentType != "" {
		SetContentTypeHeader(self.Header, self.ContentType, self.CharSet)
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
		SetLastModifiedHeader(self.Header, self.Timestamp)
	}
}
