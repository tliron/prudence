package rest

import (
	"crypto/md5"
	"hash"
	"io"

	"github.com/tliron/kutil/util"
)

//
// HashWriter
//

type HashWriter struct {
	Writer io.Writer

	hash hash.Hash
}

func NewHashWriter(writer io.Writer) *HashWriter {
	return &HashWriter{
		Writer: writer,
		hash:   md5.New(),
	}
}

func (self *HashWriter) Hash() string {
	return util.ToBase64(self.hash.Sum(nil))
}

// io.Writer
func (self *HashWriter) Write(b []byte) (int, error) {
	self.hash.Write(b)
	return self.Writer.Write(b)
}
