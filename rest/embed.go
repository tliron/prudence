package rest

import (
	"bytes"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/prudence/platform"
)

func (self *Context) Embed(present any, jsContext *commonjs.Context) error {
	if self.CacheKey != "" {
		if key, cached, ok := self.LoadCachedRepresentation(); ok {
			if len(cached.Body) == 0 {
				self.Log.Debugf("embed: ignoring cache with no body: %s", self.Request.Path)
			} else if changed, err := self.WriteCachedRepresentation(cached); err == nil {
				if changed {
					self.UpdateCachedRepresentation(key, cached)
				}
				return nil
			} else {
				return err
			}
		}
	}

	var err error
	if present, jsContext, err = commonjs.Unbind(present, jsContext); err != nil {
		return err
	}

	if self.caching() {
		buffer := bytes.NewBuffer(nil)
		writer := self.Writer
		self.Writer = buffer

		if _, err := jsContext.Environment.Call(present, self); err != nil {
			self.Writer = writer
			return err
		}

		if err := self.Flush(); err != nil {
			return err
		}

		body := buffer.Bytes()
		self.Writer = writer

		self.StoreCachedRepresentationFromBody(platform.EncodingTypeIdentity, body)
		return self.Write(body)
	} else {
		_, err := jsContext.Environment.Call(present, self)
		return err
	}
}
