package rest

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/dop251/goja"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/kutil/util"
)

//
// HandleFunc
//

type HandleFunc func(restContext *Context) (bool, error)

func GetHandleFunc(value any, jsContext *commonjs.Context) (HandleFunc, error) {
	var err error
	if value, jsContext, err = commonjs.Unbind(value, jsContext); err != nil {
		return nil, err
	}

	switch handler := value.(type) {
	case HandleFunc:
		return handler, nil

	case Handler:
		return handler.Handle, nil

	case goja.Value, commonjs.ExportedJavaScriptFunc:
		return func(restContext *Context) (bool, error) {
			if handled, err := jsContext.Environment.Call(handler, restContext); err == nil {
				if handled_, ok := handled.(bool); ok {
					return handled_, nil
				} else {
					return false, nil
				}
			} else {
				return false, err
			}
		}, nil

	default:
		return nil, fmt.Errorf("not a handler function: %T", value)
	}
}

//
// Handler
//

type Handler interface {
	Handle(restContext *Context) (bool, error) // HandleFunc signature
}

//
// NotFoundHandler
//

// ([HandleFunc] signature)
func HandleNotFound(restContext *Context) (bool, error) {
	restContext.Response.Header = make(http.Header)
	restContext.Response.Buffer = bytes.NewBuffer(util.StringToBytes("404 Not Found\n"))
	restContext.Response.Status = http.StatusNotFound
	return true, nil
}

//
// RedirectTrailingSlashHandler
//

// ([HandleFunc] signature)
func HandleRedirectTrailingSlash(restContext *Context) (bool, error) {
	return true, restContext.RedirectTrailingSlash(0)
}
