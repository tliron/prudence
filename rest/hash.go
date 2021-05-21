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
	writer io.Writer
	hash   hash.Hash
}

func NewHashWriter(writer io.Writer) *HashWriter {
	return &HashWriter{
		writer: writer,
		hash:   md5.New(),
	}
}

func (self *HashWriter) Hash() string {
	return util.ToBase64(self.hash.Sum(nil))
}

// io.Writer interface
func (self *HashWriter) Write(b []byte) (int, error) {
	self.hash.Write(b)
	return self.writer.Write(b)
}

// io.Closer interface
func (self *HashWriter) Close() error {
	return nil
}

// WrappingWriter interface
func (self *HashWriter) GetWrappedWriter() io.Writer {
	return self.writer
}
