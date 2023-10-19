package rest

import (
	"fmt"
	"net/http"
)

//
// ResponseWriter
//

// Circumvents the built-in 404 response
// See: https://stackoverflow.com/a/47286697

type ResponseWriterWrapper struct {
	response *Response
}

func (self *Response) NewResponseWriterWrapper() *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		response: self,
	}
}

// ([http.ResponseWriter] interface)
func (self *ResponseWriterWrapper) Header() http.Header {
	return self.response.Direct.Header()
}

// ([http.ResponseWriter] interface, [io.Writer] interface)
func (self *ResponseWriterWrapper) Write(p []byte) (int, error) {
	if self.response.Status != http.StatusNotFound {
		return self.response.Direct.Write(p)
	} else {
		// Don't write the 404 response but pretend that we did
		return len(p), nil
	}
}

// ([http.ResponseWriter] interface)
func (self *ResponseWriterWrapper) WriteHeader(status int) {
	fmt.Printf(">>>>>>>>>>>>>>>> %d\n", status)

	// Store status
	self.response.Status = status

	// Write all headers except 404
	if status != http.StatusNotFound {
		self.response.Direct.WriteHeader(status)
	}
}
