package rest

import (
	"bytes"
	"crypto/sha1"
	"hash"
	"io"

	"github.com/tliron/kutil/util"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag

//
// ETagBuffer
//

type ETagBuffer struct {
	Writer io.Writer
	Buffer *bytes.Buffer
	Hash   hash.Hash
}

func NewETagBuffer(writer io.Writer) *ETagBuffer {
	return &ETagBuffer{
		Writer: writer,
		Buffer: bytes.NewBuffer(nil),
		Hash:   sha1.New(),
	}
}

func (self *ETagBuffer) ETag() string {
	return util.ToBase64(self.Hash.Sum(nil))
}

// io.Writer
func (self *ETagBuffer) Write(b []byte) (int, error) {
	self.Hash.Write(b)
	return self.Buffer.Write(b)
}

// io.Closer
func (self *ETagBuffer) Close() error {
	_, err := self.Buffer.WriteTo(self.Writer)
	return err
}
