package plugin

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterAPI("myplugin", api)
}

//
// API
//

type API struct{}

var api API

func (self API) Print(message string) {
	fmt.Println(message)
}

func (self API) Badger(path string) (*badger.DB, error) {
	if db, err := badger.Open(badger.DefaultOptions(path)); err == nil {
		util.OnExit(func() {
			db.Close()
		})
		return db, nil
	} else {
		return nil, err
	}
}
