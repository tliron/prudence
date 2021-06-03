package platform

type HandleTagFunc func(context *JSTContext, code string) bool // return true to allow trailing newlines

var tagHandlers = make(map[string]HandleTagFunc)

func RegisterTag(prefix string, handle HandleTagFunc) {
	tagHandlers[prefix] = handle
}

func OnTags(f func(prefix string, handle HandleTagFunc) bool) {
	for prefix, handle := range tagHandlers {
		if !f(prefix, handle) {
			return
		}
	}
}
