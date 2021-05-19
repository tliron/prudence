package platform

type EncodeTagFunc func(context *Context, code string) bool // return true to allow trailing newlines

var tagEncoders = make(map[string]EncodeTagFunc)

func RegisterTag(prefix string, encodeTag EncodeTagFunc) {
	tagEncoders[prefix] = encodeTag
}

func OnTags(f func(prefix string, encodeTag EncodeTagFunc) bool) {
	for prefix, encodeTag := range tagEncoders {
		if !f(prefix, encodeTag) {
			return
		}
	}
}
