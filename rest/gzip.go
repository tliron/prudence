package rest

import (
	"io"

	"github.com/valyala/fasthttp"
)

//
// GZipWriter
//

type GZipWriter struct {
	Writer io.Writer
}

func NewGZipWriter(writer io.Writer) *GZipWriter {
	return &GZipWriter{writer}
}

// io.Writer
func (self *GZipWriter) Write(b []byte) (int, error) {
	return fasthttp.WriteGzip(self.Writer, b)
}
