package rest

import (
	"bytes"
	"io"

	"github.com/valyala/fasthttp"
)

//
// EncodingType
//

type EncodingType int

const (
	EncodingTypeUnsupported = EncodingType(-1)
	EncodingTypePlain       = EncodingType(0)
	EncodingTypeBrotli      = EncodingType(1)
	EncodingTypeGZip        = EncodingType(2)
	EncodingTypeDeflate     = EncodingType(3)
)

// fmt.Stringer interface
func (self EncodingType) String() string {
	switch self {
	case EncodingTypePlain:
		return "plain"
	case EncodingTypeBrotli:
		return "brotli"
	case EncodingTypeGZip:
		return "gzip"
	case EncodingTypeDeflate:
		return "deflate"
	default:
		return "unsupported"
	}
}

func GetEncodingType(name string) EncodingType {
	switch name {
	case "":
		return EncodingTypePlain
	case "br":
		return EncodingTypeBrotli
	case "gzip":
		return EncodingTypeGZip
	case "deflate":
		return EncodingTypeDeflate
	default:
		return EncodingTypeUnsupported
	}
}

//
// EncodeWriter
//

type EncodeWriter struct {
	Writer io.Writer
	Type   EncodingType

	buffer *bytes.Buffer
}

func NewEncodeWriter(writer io.Writer, type_ EncodingType) *EncodeWriter {
	return &EncodeWriter{
		Writer: writer,
		Type:   type_,
		buffer: bytes.NewBuffer(nil),
	}
}

func SetBestEncodeWriter(context *Context) {
	if context.context.Request.Header.HasAcceptEncoding("br") {
		AddContentEncoding(context.context, "br")
		context.writer = NewEncodeWriter(context.writer, EncodingTypeBrotli)
	} else if context.context.Request.Header.HasAcceptEncoding("gzip") {
		AddContentEncoding(context.context, "gzip")
		context.writer = NewEncodeWriter(context.writer, EncodingTypeGZip)
	} else if context.context.Request.Header.HasAcceptEncoding("deflate") {
		AddContentEncoding(context.context, "deflate")
		context.writer = NewEncodeWriter(context.writer, EncodingTypeDeflate)
	}
}

// io.Writer
func (self *EncodeWriter) Write(b []byte) (int, error) {
	return self.buffer.Write(b)
}

// io.Close
func (self *EncodeWriter) Close() error {
	switch self.Type {
	case EncodingTypeBrotli:
		_, err := fasthttp.WriteBrotli(self.Writer, self.buffer.Bytes())
		return err

	case EncodingTypeGZip:
		_, err := fasthttp.WriteGzip(self.Writer, self.buffer.Bytes())
		return err

	case EncodingTypeDeflate:
		_, err := fasthttp.WriteDeflate(self.Writer, self.buffer.Bytes())
		return err

	default:
		return nil
	}
}
