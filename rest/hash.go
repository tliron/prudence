package rest

import (
	"crypto/md5"
	"errors"
	"hash"
	"io"

	"github.com/tliron/kutil/util"
)

// Calculating a signature from the body is not that great. It saves bandwidth but not computing
// resources, as we still need to generate the body in order to calculate the signature. Ideally,
// the signature should be based on the data sources used to generate the page.
//
// https://www.mnot.net/blog/2007/08/07/etags
// http://www.tbray.org/ongoing/When/200x/2007/07/31/Design-for-the-Web
func (self *Context) StartSignature() {
	if _, ok := self.writer.(*HashWriter); !ok {
		self.writer = NewHashWriter(self.writer)
	}
}

func (self *Context) EndSignature() error {
	if hashWriter, ok := self.writer.(*HashWriter); ok {
		self.Response.Signature = hashWriter.Hash()
		self.writer = hashWriter.writer
		return nil
	} else {
		return errors.New("did not call startSignature()")
	}
}

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
