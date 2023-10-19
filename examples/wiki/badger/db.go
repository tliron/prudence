package badger

import (
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/kutil/util"
)

//
// DB
//

type DB struct {
	api      *BadgerAPI
	db       *badger.DB
	gcTicker *time.Ticker
}

func (self *BadgerAPI) NewDB(db *badger.DB) *DB {
	return &DB{
		api: self,
		db:  db,
	}
}

func (self *BadgerAPI) Open(path string, options any) (*DB, error) {
	if options_, err := NewDBOptions(path, options); err == nil {
		if db, err := badger.Open(options_); err == nil {
			util.OnExitError(db.Close)
			db_ := self.NewDB(db)
			db_.startGc()
			return db_, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *DB) View(f any) error {
	return self.db.View(func(txn *badger.Txn) error {
		_, err := commonjs.Call(self.api.jsContext.Environment.Runtime, f, nil, self.api.NewTxn(txn))
		return err
	})
}

func (self *DB) Update(f any) error {
	return self.db.Update(func(txn *badger.Txn) error {
		_, err := commonjs.Call(self.api.jsContext.Environment.Runtime, f, nil, self.api.NewTxn(txn))
		return err
	})
}

func (self *DB) GetSequence(key any, bandwidth uint64) (*badger.Sequence, error) {
	return self.db.GetSequence(util.ToBytes(key), bandwidth)
}

func (self *DB) startGc() {
	self.gcTicker = time.NewTicker(5 * time.Minute)
	util.OnExit(self.gcTicker.Stop)
	go self.gc()
}

func (self *DB) gc() {
	// https://dgraph.io/docs/badger/get-started/#garbage-collection
	for range self.gcTicker.C {
	again:
		log.Debug("gc")
		if err := self.db.RunValueLogGC(0.7); err == nil {
			goto again
		}
	}
}
