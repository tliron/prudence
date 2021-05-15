package rest

import (
	"bytes"
	"io"

	"github.com/valyala/fasthttp"
)

//
// EncodeWriter
//

type EncodeWriter struct {
	Writer   io.Writer
	Encoding string

	buffer *bytes.Buffer
}

func NewEncodeWriter(writer io.Writer, encoding string) *EncodeWriter {
	return &EncodeWriter{
		Writer:   writer,
		Encoding: encoding,
		buffer:   bytes.NewBuffer(nil),
	}
}

// io.Writer
func (self *EncodeWriter) Write(b []byte) (int, error) {
	return self.buffer.Write(b)
}

// io.Close
func (self *EncodeWriter) Close() error {
	switch self.Encoding {
	case "gzip":
		_, err := fasthttp.WriteGzip(self.Writer, self.buffer.Bytes())
		return err

	case "deflate":
		_, err := fasthttp.WriteDeflate(self.Writer, self.buffer.Bytes())
		return err

	case "br":
		_, err := fasthttp.WriteBrotli(self.Writer, self.buffer.Bytes())
		return err

	default:
		return nil
	}
}
