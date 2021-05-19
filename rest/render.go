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
	Writer         io.Writer
	Render         platform.RenderFunc
	GetRelativeURL platform.GetRelativeURL

	buffer *bytes.Buffer
}

func NewRenderWriter(writer io.Writer, renderer string, getRelativeURL platform.GetRelativeURL) (*RenderWriter, error) {
	if render_, err := platform.GetRenderer(renderer); err == nil {
		// Note: renderer can be nil
		return &RenderWriter{
			Writer:         writer,
			Render:         render_,
			GetRelativeURL: getRelativeURL,
			buffer:         bytes.NewBuffer(nil),
		}, nil
	} else {
		return nil, err
	}
}

// io.Writer
func (self *RenderWriter) Write(b []byte) (int, error) {
	if self.Render == nil {
		// Optimize for empty renderer
		return self.Writer.Write(b)
	} else {
		return self.buffer.Write(b)
	}
}

// io.Close
func (self *RenderWriter) Close() error {
	if self.Render == nil {
		// Optimize for empty renderer
		return nil
	} else if content, err := self.Render(util.BytesToString(self.buffer.Bytes()), self.GetRelativeURL); err == nil {
		_, err = self.Writer.Write(util.StringToBytes(content))
		return err
	} else {
		return err
	}
}
