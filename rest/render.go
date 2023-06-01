package rest

import (
	"bytes"
	"errors"
	"io"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

func (self *Context) StartRender(renderer string, jsContext *commonjs.Context) error {
	if renderWriter, err := NewRenderWriter(self.writer, renderer, jsContext); err == nil {
		self.writer = renderWriter
		return nil
	} else {
		return err
	}
}

func (self *Context) EndRender() error {
	if renderWriter, ok := self.writer.(*RenderWriter); ok {
		err := renderWriter.Close()
		self.writer = renderWriter.GetWrappedWriter()
		return err
	} else {
		return errors.New("did not call startRender()")
	}
}

//
// RenderWriter
//

type RenderWriter struct {
	writer  io.Writer
	render  platform.RenderFunc
	context *commonjs.Context
	buffer  *bytes.Buffer
}

func NewRenderWriter(writer io.Writer, renderer string, context *commonjs.Context) (*RenderWriter, error) {
	if render_, err := platform.GetRenderer(renderer); err == nil {
		// Note: renderer can be nil
		return &RenderWriter{
			writer:  writer,
			render:  render_,
			context: context,
			buffer:  bytes.NewBuffer(nil),
		}, nil
	} else {
		return nil, err
	}
}

// io.Writer interface
func (self *RenderWriter) Write(b []byte) (int, error) {
	if self.render == nil {
		// Optimize for empty renderer
		return self.writer.Write(b)
	} else {
		return self.buffer.Write(b)
	}
}

// io.Closer interface
func (self *RenderWriter) Close() error {
	if self.render == nil {
		// Optimize for empty renderer
		return nil
	} else if content, err := self.render(util.BytesToString(self.buffer.Bytes()), self.context); err == nil {
		_, err = self.writer.Write(util.StringToBytes(content))
		return err
	} else {
		return err
	}
}

// WrappingWriter interface
func (self *RenderWriter) GetWrappedWriter() io.Writer {
	return self.writer
}
