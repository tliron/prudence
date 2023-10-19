package badger

import (
	"bytes"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

//
// Txn
//

type Txn struct {
	api *BadgerAPI
	txn *badger.Txn
}

func (self *BadgerAPI) NewTxn(txn *badger.Txn) *Txn {
	return &Txn{self, txn}
}

func (self *Txn) Get(key any) (*Item, error) {
	if item, err := self.txn.Get(util.ToBytes(key)); err == nil {
		return self.api.NewItem(item), nil
	} else {
		return nil, err
	}
}

func (self *Txn) Set(key any, value any, format string) error {
	if format == "" {
		format = self.api.DefaultFormat
	}

	var value_ []byte

	switch format {
	case "", "bytes", "string":
		value_ = util.ToBytes(value)

	default:
		var buffer bytes.Buffer
		transcriber := transcribe.Transcriber{
			Writer: &buffer,
			Format: format,
		}

		if err := transcriber.Write(value); err != nil {
			return err
		}

		value_ = buffer.Bytes()
	}

	return self.txn.Set(util.ToBytes(key), value_)
}

func (self *Txn) Iterate(f any, options any) error {
	options_, err := NewIteratorOptions(options)
	if err != nil {
		return err
	}

	iterator := self.txn.NewIterator(options_)
	defer iterator.Close()

	for iterator.Rewind(); iterator.Valid(); iterator.Next() {
		item := iterator.Item()

		if r, err := commonjs.Call(self.api.jsContext.Environment.Runtime, f, nil, self.api.NewItem(item)); err == nil {
			if r_, ok := r.(bool); ok {
				if !r_ {
					break
				}
			}
		} else {
			return err
		}
	}

	return nil
}
