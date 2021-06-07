package rest

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/tliron/prudence/platform"
)

func GetEncodingType(name string) platform.EncodingType {
	switch name {
	case "identity", "":
		return platform.EncodingTypeIdentity
	case "br":
		return platform.EncodingTypeBrotli
	case "gzip":
		return platform.EncodingTypeGZip
	case "deflate":
		return platform.EncodingTypeFlate
	default:
		return platform.EncodingTypeUnsupported
	}
}

func NegotiateBestEncodingType(header http.Header) platform.EncodingType {
	clientEncodings := strings.Split(header.Get(HeaderAcceptEncoding), ",")
	for _, clientEncoding := range clientEncodings {
		if type_ := GetEncodingType(clientEncoding); type_ != platform.EncodingTypeUnsupported {
			return type_
		}
	}

	return platform.EncodingTypeIdentity
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
	type_ := NegotiateBestEncodingType(context.Request.Header)
	switch type_ {
	case platform.EncodingTypeBrotli:
		context.Response.Header.Set(HeaderContentEncoding, "br")
		context.writer = NewEncodeWriter(context.writer, type_)
	case platform.EncodingTypeGZip:
		context.Response.Header.Set(HeaderContentEncoding, "gzip")
		context.writer = NewEncodeWriter(context.writer, type_)
	case platform.EncodingTypeFlate:
		context.Response.Header.Set(HeaderContentEncoding, "deflate")
		context.writer = NewEncodeWriter(context.writer, type_)
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
		return platform.EncodeBrotli(self.buffer.Bytes(), self.writer)

	case platform.EncodingTypeGZip:
		return platform.EncodeGZip(self.buffer.Bytes(), self.writer)

	case platform.EncodingTypeFlate:
		return platform.EncodeFlate(self.buffer.Bytes(), self.writer)

	default:
		return nil
	}
}

// WrappingWriter interface
func (self *EncodeWriter) GetWrappedWriter() io.Writer {
	return self.writer
}
