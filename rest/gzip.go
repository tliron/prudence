package rest

import (
	"bytes"
	"io"

	"github.com/valyala/fasthttp"
)

//
// GZipWriter
//

type GZipWriter struct {
	Writer io.Writer

	buffer *bytes.Buffer
}

func NewGZipWriter(writer io.Writer) *GZipWriter {
	return &GZipWriter{
		Writer: writer,
		buffer: bytes.NewBuffer(nil),
	}
}

// io.Writer
func (self *GZipWriter) Write(b []byte) (int, error) {
	return self.buffer.Write(b)
}

// io.Close
func (self *GZipWriter) Close() error {
	_, err := fasthttp.WriteGzip(self.Writer, self.buffer.Bytes())
	return err
}
