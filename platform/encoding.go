package platform

import (
	bytespkg "bytes"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
	"github.com/tliron/commonlog"
)

//
// EncodingType
//

type EncodingType int

const (
	EncodingTypeUnsupported = EncodingType(-1)
	EncodingTypeIdentity    = EncodingType(0)

	EncodingTypeBrotli    = EncodingType(1)
	EncodingTypeCompress  = EncodingType(2)
	EncodingTypeDeflate   = EncodingType(3)
	EncodingTypeGZip      = EncodingType(4)
	EncodingTypeZstandard = EncodingType(5)
)

// ([fmt.Stringer] interface)
func (self EncodingType) String() string {
	switch self {
	case EncodingTypeIdentity:
		return "Identity"
	case EncodingTypeBrotli:
		return "Brotli"
	case EncodingTypeCompress:
		return "Compress"
	case EncodingTypeDeflate:
		return "Deflate"
	case EncodingTypeGZip:
		return "GZip"
	case EncodingTypeZstandard:
		return "Zstandard"
	default:
		return "Unsupported"
	}
}

func (self EncodingType) Header() string {
	switch self {
	case EncodingTypeBrotli:
		return "br"
	case EncodingTypeCompress:
		return "compress"
	case EncodingTypeDeflate:
		return "deflate"
	case EncodingTypeGZip:
		return "gzip"
	case EncodingTypeZstandard:
		return "zstd"
	default:
		return ""
	}
}

func GetEncodingFromHeader(name string) EncodingType {
	switch name {
	case "":
		return EncodingTypeIdentity
	case "br":
		return EncodingTypeBrotli
	case "compress":
		return EncodingTypeCompress
	case "deflate":
		return EncodingTypeDeflate
	case "gzip":
		return EncodingTypeGZip
	case "zstd":
		return EncodingTypeZstandard
	default:
		return EncodingTypeUnsupported
	}
}

func (self EncodingType) Encode(bytes []byte, writer io.Writer) error {
	switch self {
	case EncodingTypeBrotli:
		return EncodeBrotli(bytes, writer)
	case EncodingTypeDeflate:
		return EncodeDeflate(bytes, writer)
	case EncodingTypeGZip:
		return EncodeGZip(bytes, writer)
	case EncodingTypeZstandard:
		return EncodeZstandard(bytes, writer)
	default:
		return fmt.Errorf("unsupported encoding: %d", self)
	}
}

func (self EncodingType) Decode(bytes []byte, writer io.Writer) error {
	switch self {
	case EncodingTypeBrotli:
		return DecodeBrotli(bytes, writer)
	case EncodingTypeDeflate:
		return DecodeDeflate(bytes, writer)
	case EncodingTypeGZip:
		return DecodeGZip(bytes, writer)
	case EncodingTypeZstandard:
		return DecodeZstandard(bytes, writer)
	default:
		return fmt.Errorf("unsupported encoding: %d", self)
	}
}

func (self EncodingType) Encoded(bytes []byte) ([]byte, error) {
	if self == EncodingTypeIdentity {
		return bytes, nil
	}

	buffer := bytespkg.NewBuffer(nil)
	if err := self.Encode(bytes, buffer); err == nil {
		return buffer.Bytes(), nil
	} else {
		return nil, err
	}
}

func (self EncodingType) Decoded(bytes []byte) ([]byte, error) {
	if self == EncodingTypeIdentity {
		return bytes, nil
	}

	buffer := bytespkg.NewBuffer(nil)
	if err := self.Decode(bytes, buffer); err == nil {
		return buffer.Bytes(), nil
	} else {
		return nil, err
	}
}

func EncodeBrotli(bytes []byte, writer io.Writer) error {
	writer_ := brotli.NewWriter(writer)
	defer commonlog.CallAndLogWarning(writer_.Close, "EncodeBrotli.Close", log)
	_, err := writer_.Write(bytes)
	return err
}

func DecodeBrotli(bytes []byte, writer io.Writer) error {
	reader := brotli.NewReader(bytespkg.NewReader(bytes))
	_, err := io.Copy(writer, reader)
	return err
}

func EncodeDeflate(bytes []byte, writer io.Writer) error {
	writer_ := zlib.NewWriter(writer)
	defer commonlog.CallAndLogWarning(writer_.Close, "EncodeDeflate.Close", log)
	_, err := writer_.Write(bytes)
	return err
}

func DecodeDeflate(bytes []byte, writer io.Writer) error {
	if reader, err := zlib.NewReader(bytespkg.NewReader(bytes)); err == nil {
		_, err = io.Copy(writer, reader)
		return err
	} else {
		return err
	}
}

func EncodeGZip(bytes []byte, writer io.Writer) error {
	writer_ := pgzip.NewWriter(writer)
	defer commonlog.CallAndLogWarning(writer_.Close, "EncodeGZip.Close", log)
	_, err := writer_.Write(bytes)
	return err
}

func DecodeGZip(bytes []byte, writer io.Writer) error {
	if reader, err := pgzip.NewReader(bytespkg.NewReader(bytes)); err == nil {
		_, err := io.Copy(writer, reader)
		return err
	} else {
		return err
	}
}

func EncodeZstandard(bytes []byte, writer io.Writer) error {
	if writer_, err := zstd.NewWriter(writer); err == nil {
		defer commonlog.CallAndLogWarning(writer_.Close, "EncodeZstandard.Close", log)
		_, err := writer_.Write(bytes)
		return err
	} else {
		return err
	}
}

func DecodeZstandard(bytes []byte, writer io.Writer) error {
	if reader, err := zstd.NewReader(bytespkg.NewReader(bytes)); err == nil {
		_, err := io.Copy(writer, reader)
		return err
	} else {
		return err
	}
}
