package platform

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/andybalholm/brotli"
)

//
// EncodingType
//

type EncodingType int

const (
	EncodingTypeUnsupported = EncodingType(-1)
	EncodingTypeIdentity    = EncodingType(0)
	EncodingTypeBrotli      = EncodingType(1)
	EncodingTypeGZip        = EncodingType(2)
	EncodingTypeFlate       = EncodingType(3)
)

// fmt.Stringer interface
func (self EncodingType) String() string {
	switch self {
	case EncodingTypeIdentity:
		return "identity"
	case EncodingTypeBrotli:
		return "brotli"
	case EncodingTypeGZip:
		return "gzip"
	case EncodingTypeFlate:
		return "flate"
	default:
		return "unsupported"
	}
}

func EncodeBrotli(bytes []byte, writer io.Writer) error {
	writer_ := brotli.NewWriter(writer)
	defer writer_.Close()
	_, err := writer_.Write(bytes)
	return err
}

func DecodeBrotli(bytes_ []byte, writer io.Writer) error {
	reader := brotli.NewReader(bytes.NewReader(bytes_))
	_, err := io.Copy(writer, reader)
	return err
}

func EncodeGZip(bytes []byte, writer io.Writer) error {
	writer_ := gzip.NewWriter(writer)
	defer writer_.Close()
	_, err := writer_.Write(bytes)
	return err
}

func DecodeGZip(bytes_ []byte, writer io.Writer) error {
	if reader, err := gzip.NewReader(bytes.NewReader(bytes_)); err == nil {
		_, err := io.Copy(writer, reader)
		return err
	} else {
		return err
	}
}

func EncodeFlate(bytes []byte, writer io.Writer) error {
	writer_ := zlib.NewWriter(writer)
	defer writer_.Close()
	_, err := writer_.Write(bytes)
	return err
}

func DecodeFlate(bytes_ []byte, writer io.Writer) error {
	if reader, err := zlib.NewReader(bytes.NewReader(bytes_)); err == nil {
		_, err := io.Copy(writer, reader)
		return err
	} else {
		return err
	}
}
