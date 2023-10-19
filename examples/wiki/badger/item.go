package badger

import (
	"bytes"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

//
// Item
//

type Item struct {
	api  *BadgerAPI
	item *badger.Item
}

func (self *BadgerAPI) NewItem(item *badger.Item) *Item {
	return &Item{self, item}
}

func (self *Item) Key() string {
	return util.BytesToString(self.item.Key())
}

func (self *Item) Value(f any, format string) error {
	if format == "" {
		format = self.api.DefaultFormat
	}

	return self.item.Value(func(value []byte) error {
		var value_ any
		switch format {
		case "", "bytes":
			value_ = value

		case "string":
			value_ = util.BytesToString(value)

		default:
			var err error
			value_, _, err = ard.Read(bytes.NewReader(value), format, false)
			if err != nil {
				return err
			}

			value_, _ = ard.ConvertMapsToStringMaps(value_)
		}

		_, err := commonjs.Call(self.api.jsContext.Environment.Runtime, f, nil, value_)
		return err
	})
}
