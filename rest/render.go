package rest

import (
	"bytes"
	"io"

	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

//
// RenderWriter
//

type RenderWriter struct {
	writer         io.Writer
	render         platform.RenderFunc
	getRelativeURL platform.GetRelativeURL
	buffer         *bytes.Buffer
}

func NewRenderWriter(writer io.Writer, renderer string, getRelativeURL platform.GetRelativeURL) (*RenderWriter, error) {
	if render_, err := platform.GetRenderer(renderer); err == nil {
		// Note: renderer can be nil
		return &RenderWriter{
			writer:         writer,
			render:         render_,
			getRelativeURL: getRelativeURL,
			buffer:         bytes.NewBuffer(nil),
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
	} else if content, err := self.render(util.BytesToString(self.buffer.Bytes()), self.getRelativeURL); err == nil {
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
