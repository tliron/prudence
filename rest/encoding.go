package rest

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/tliron/prudence/platform"
)

//
// EncodingPreference
//

type EncodingPreference struct {
	Name   string
	Type   platform.EncodingType
	Weight float64
}

func ParseEncodingPreference(text string) (EncodingPreference, error) {
	self := EncodingPreference{Weight: 1.0}

	s := strings.SplitN(text, ";", 2)
	self.Name = s[0]
	self.Type = GetEncodingType(self.Name)

	// Annotation
	if len(s) == 2 {
		annotationText := s[1]
		if strings.HasPrefix(annotationText, "q=") {
			var err error
			if self.Weight, err = strconv.ParseFloat(annotationText[2:], 64); err != nil {
				return self, err
			}
		}
	}

	return self, nil
}

// fmt.Stringify interface
func (self EncodingPreference) String() string {
	return fmt.Sprintf("%s;q=%g", self.Name, self.Weight)
}

//
// EncodingPreferences
//

type EncodingPreferences []EncodingPreference

func ParseEncodingPreferences(text string) EncodingPreferences {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Encoding

	var self EncodingPreferences

	if text = strings.TrimSpace(text); text != "" {
		for _, text_ := range strings.Split(text, ",") {
			text_ = strings.TrimSpace(text_)
			if encodingPreference, err := ParseEncodingPreference(text_); err == nil {
				self = append(self, encodingPreference)
			}
		}

		sort.Stable(sort.Reverse(self))
	}

	//log.Infof("%s", text)
	//log.Infof("%v", self)

	return self
}

func (self EncodingPreferences) NegotiateBest(context *Context) platform.EncodingType {
	for _, encodingPreference := range self {
		if encodingPreference.Type != platform.EncodingTypeUnsupported {
			return encodingPreference.Type
		}
	}

	return platform.EncodingTypeIdentity
}

// sort.Interface interface
func (self EncodingPreferences) Len() int {
	return len(self)
}

// sort.Interface interface
func (self EncodingPreferences) Less(i int, j int) bool {
	return self[i].Weight < self[j].Weight
}

// sort.Interface interface
func (self EncodingPreferences) Swap(i int, j int) {
	self[i], self[j] = self[j], self[i]
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
	encodingPreferences := ParseEncodingPreferences(context.Request.Header.Get(HeaderAcceptEncoding))
	type_ := encodingPreferences.NegotiateBest(context)
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

// Utils

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
