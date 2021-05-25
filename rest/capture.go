package rest

import (
	"bytes"
	"io"
)

type CaptureFunc func(name string, value string)

//
// CaptureWriter
//

type CaptureWriter struct {
	writer  io.Writer
	name    string
	capture CaptureFunc
	buffer  *bytes.Buffer
}

func NewCaptureWriter(writer io.Writer, name string, capture CaptureFunc) *CaptureWriter {
	return &CaptureWriter{
		writer:  writer,
		name:    name,
		capture: capture,
		buffer:  bytes.NewBuffer(nil),
	}
}

// io.Writer interface
func (self *CaptureWriter) Write(b []byte) (int, error) {
	return self.buffer.Write(b)
}

// io.Closer interface
func (self *CaptureWriter) Close() error {
	self.capture(self.name, self.buffer.String())
	return nil
}

// WrappingWriter interface
func (self *CaptureWriter) GetWrappedWriter() io.Writer {
	return self.writer
}
