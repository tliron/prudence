package badger

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterExtension("badger", func(jsContext *commonjs.Context) any {
		return NewBadgerAPI(jsContext)
	})
}

//
// BadgerAPI
//

type BadgerAPI struct {
	DefaultFormat string

	jsContext *commonjs.Context
}

func NewBadgerAPI(jsContext *commonjs.Context) *BadgerAPI {
	return &BadgerAPI{
		DefaultFormat: "cbor",
		jsContext:     jsContext,
	}
}
