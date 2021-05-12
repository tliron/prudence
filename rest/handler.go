package rest

//
// HandleFunc
//

type HandleFunc func(context *Context) bool

// HandleFunc signature
func Handled(context *Context) bool {
	return true
}

//
// Handler
//

type Handler interface {
	Handle(context *Context) bool
}

func GetHandleFunc(value interface{}) (HandleFunc, bool) {
	if handler, ok := value.(Handler); ok {
		return handler.Handle, true
	} else {
		return nil, false
	}
}

var DefaultNotFound = &defaultNotFound{}

type defaultNotFound struct{}

// Handler interface
// HandleFunc signature
func (self *defaultNotFound) Handle(context *Context) bool {
	context.Context.NotFound()
	return true
}
