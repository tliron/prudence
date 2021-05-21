package rest

import (
	"bytes"
	"io"

	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

func GetEncodingType(name string) platform.EncodingType {
	switch name {
	case "":
		return platform.EncodingTypePlain
	case "br":
		return platform.EncodingTypeBrotli
	case "gzip":
		return platform.EncodingTypeGZip
	case "deflate":
		return platform.EncodingTypeDeflate
	default:
		return platform.EncodingTypeUnsupported
	}
}

//
// EncodeWriter
//

type EncodeWriter struct {
	writer io.Writer
	type_  platform.EncodingType
	buffer *bytes.Buffer
}

func NewEncodeWriter(writer io.Writer, type_ platform.EncodingType) *EncodeWriter {
	return &EncodeWriter{
		writer: writer,
		type_:  type_,
		buffer: bytes.NewBuffer(nil),
	}
}

func SetBestEncodeWriter(context *Context) {
	if context.context.Request.Header.HasAcceptEncoding("br") {
		AddContentEncoding(context.context, "br")
		context.writer = NewEncodeWriter(context.writer, platform.EncodingTypeBrotli)
	} else if context.context.Request.Header.HasAcceptEncoding("gzip") {
		AddContentEncoding(context.context, "gzip")
		context.writer = NewEncodeWriter(context.writer, platform.EncodingTypeGZip)
	} else if context.context.Request.Header.HasAcceptEncoding("deflate") {
		AddContentEncoding(context.context, "deflate")
		context.writer = NewEncodeWriter(context.writer, platform.EncodingTypeDeflate)
	}
}

// io.Writer interface
func (self *EncodeWriter) Write(b []byte) (int, error) {
	return self.buffer.Write(b)
}

// io.Closer interface
func (self *EncodeWriter) Close() error {
	switch self.type_ {
	case platform.EncodingTypeBrotli:
		_, err := fasthttp.WriteBrotli(self.writer, self.buffer.Bytes())
		return err

	case platform.EncodingTypeGZip:
		_, err := fasthttp.WriteGzip(self.writer, self.buffer.Bytes())
		return err

	case platform.EncodingTypeDeflate:
		_, err := fasthttp.WriteDeflate(self.writer, self.buffer.Bytes())
		return err

	default:
		return nil
	}
}

// WrappingWriter interface
func (self *EncodeWriter) GetWrappedWriter() io.Writer {
	return self.writer
}
