package platform

import (
	"bytes"
	"io"
)

//
// EncodeWriter
//

type EncodeWriter struct {
	encoding EncodingType
	writer   io.Writer
	buffer   *bytes.Buffer
}

func (self EncodingType) NewWriter(writer io.Writer) *EncodeWriter {
	return &EncodeWriter{
		encoding: self,
		writer:   writer,
		buffer:   bytes.NewBuffer(nil),
	}
}

// ([io.Writer] interface)
func (self *EncodeWriter) Write(b []byte) (int, error) {
	return self.buffer.Write(b)
}

// [io.StringWriter] interface
func (self *EncodeWriter) WriteString(s string) (int, error) {
	return self.buffer.WriteString(s)
}

// [io.ByteWriter] interface
func (self *EncodeWriter) WriteByte(c byte) error {
	return self.buffer.WriteByte(c)
}

// ([io.Closer] interface)
func (self *EncodeWriter) Close() error {
	return self.encoding.Encode(self.buffer.Bytes(), self.writer)
}

// ([jst.WrappingWriter] interface)
func (self *EncodeWriter) GetWrappedWriter() io.Writer {
	return self.writer
}
