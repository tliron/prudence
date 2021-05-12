package rest

import (
	"bytes"
	"crypto/sha1"
	"hash"
	"io"

	"github.com/tliron/kutil/util"
)

//
// HashWriter
//

type HashWriter struct {
	Writer io.Writer

	buffer *bytes.Buffer
	hash   hash.Hash
}

func NewHashWriter(writer io.Writer) *HashWriter {
	return &HashWriter{
		Writer: writer,
		buffer: bytes.NewBuffer(nil),
		hash:   sha1.New(),
	}
}

func (self *HashWriter) Hash() string {
	return util.ToBase64(self.hash.Sum(nil))
}

// io.Writer
func (self *HashWriter) Write(b []byte) (int, error) {
	self.hash.Write(b)
	return self.buffer.Write(b)
}

// io.Closer
func (self *HashWriter) Close() error {
	_, err := self.buffer.WriteTo(self.Writer)
	return err
}
