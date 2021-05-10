package rest

//
// HandlerFunc
//

type HandlerFunc func(context *Context) bool

// HandlerFunc signature
func Handled(context *Context) bool {
	return true
}

//
// Handler
//

type Handler interface {
	Handle(context *Context) bool
}

func GetHandler(value interface{}) (HandlerFunc, bool) {
	if handler, ok := value.(Handler); ok {
		return handler.Handle, true
	} else {
		return nil, false
	}
}

var DefaultNotFound = &defaultNotFound{}

type defaultNotFound struct{}

// Handler interface
// HandlerFunc signature
func (self *defaultNotFound) Handle(context *Context) bool {
	context.RequestContext.NotFound()
	return true
}
