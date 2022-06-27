package rest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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

func (self EncodingPreferences) ForbidIdentity() bool {
	for _, encodingPreference := range self {
		if (encodingPreference.Type == platform.EncodingTypeIdentity) || (encodingPreference.Name == "*") {
			if encodingPreference.Weight == 0.0 {
				return true
			}
		}
	}

	return false
}

func (self EncodingPreferences) NegotiateBest(context *Context) platform.EncodingType {
	for _, encodingPreference := range self {
		if encodingPreference.Weight != 0.0 {
			switch encodingPreference.Type {
			// Note: "compress" has been deprecated
			case platform.EncodingTypeUnsupported, platform.EncodingTypeCompress:
			default:
				return encodingPreference.Type
			}
		}
	}

	if !self.ForbidIdentity() {
		return platform.EncodingTypeIdentity
	} else {
		return platform.EncodingTypeUnsupported
	}
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

func SetBestEncodeWriter(context *Context) bool {
	encodingPreferences := ParseEncodingPreferences(context.Request.Header.Get(HeaderAcceptEncoding))
	type_ := encodingPreferences.NegotiateBest(context)
	switch type_ {
	case platform.EncodingTypeIdentity:
		return true

	case platform.EncodingTypeBrotli:
		context.Response.Header.Set(HeaderContentEncoding, "br")
		context.writer = NewEncodeWriter(context.writer, type_)
		return true

	case platform.EncodingTypeDeflate:
		context.Response.Header.Set(HeaderContentEncoding, "deflate")
		context.writer = NewEncodeWriter(context.writer, type_)
		return true

	case platform.EncodingTypeGZip:
		context.Response.Header.Set(HeaderContentEncoding, "gzip")
		context.writer = NewEncodeWriter(context.writer, type_)
		return true

	default:
		context.Response.Status = http.StatusNotAcceptable
		return false
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

	case platform.EncodingTypeDeflate:
		return platform.EncodeDeflate(self.buffer.Bytes(), self.writer)

	case platform.EncodingTypeGZip:
		return platform.EncodeGZip(self.buffer.Bytes(), self.writer)

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
	case "compress":
		return platform.EncodingTypeCompress
	case "deflate":
		return platform.EncodingTypeDeflate
	case "gzip":
		return platform.EncodingTypeGZip
	default:
		return platform.EncodingTypeUnsupported
	}
}
