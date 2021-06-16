package rest

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/util"
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

func GetHandleFunc(value interface{}, jsContext *js.Context) (HandleFunc, error) {
	if bind, ok := value.(js.Bind); ok {
		var err error
		if value, jsContext, err = bind.Unbind(); err != nil {
			return nil, err
		}
	}

	if handler, ok := value.(Handler); ok {
		return handler.Handle, nil
	} else if function, ok := value.(js.JavaScriptFunc); ok {
		return func(context *Context) bool {
			handled := jsContext.Environment.Call(function, context)
			if handled_, ok := handled.(bool); ok {
				return handled_
			} else {
				return false
			}
		}, nil
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
	context.Response.Buffer = bytes.NewBuffer(util.StringToBytes("404 Not Found\n"))
	context.Response.Status = http.StatusNotFound
	return true
}
