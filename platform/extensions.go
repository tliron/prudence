package platform

import (
	"github.com/tliron/commonjs-goja"
)

var Extensions = make(map[string]commonjs.CreateExtensionFunc)

func RegisterExtension(name string, extension commonjs.CreateExtensionFunc) {
	Extensions[name] = extension
}
