package platform

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
