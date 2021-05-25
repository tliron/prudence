package rest

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

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

func GetHandleFunc(value interface{}) (HandleFunc, error) {
	if handler, ok := value.(Handler); ok {
		return handler.Handle, nil
	} else {
		return nil, fmt.Errorf("not a handler: %T", value)
	}
}

//
// DefaultNotFound
//

var DefaultNotFound defaultNotFound

type defaultNotFound struct{}

// Handler interface
// HandleFunc signature
func (self defaultNotFound) Handle(context *Context) bool {
	context.Context.Response.Reset()
	context.Context.SetStatusCode(fasthttp.StatusNotFound)
	context.Context.SetBodyString("404 Not Found\n")
	return true
}
