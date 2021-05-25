package rest

import (
	"github.com/dop251/goja"
)

type JavaScriptFunc = func(goja.FunctionCall) goja.Value

func CallJavaScript(runtime *goja.Runtime, function JavaScriptFunc, arguments ...interface{}) interface{} {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("%s", r)
		}
	}()

	arguments_ := make([]goja.Value, len(arguments))
	for index, argument := range arguments {
		arguments_[index] = runtime.ToValue(argument)
	}

	return function(goja.FunctionCall{
		This:      nil,
		Arguments: arguments_,
	}).Export()
}
