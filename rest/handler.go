package rest

import (
	"bytes"
	"fmt"
	"net/http"

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
	context.Response.Buffer = bytes.NewBuffer(util.StringToBytes("404 Not Found\n"))
	context.Response.Status = http.StatusNotFound
	return true
}
