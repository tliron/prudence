package rest

import (
	"fmt"
	"io"
	"time"

	"github.com/tliron/kutil/js"
	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

//
// RepresentionFunc
//

type RepresentionFunc func(context *Context) error

func NewRepresentationFunc(hook *js.Hook) RepresentionFunc {
	if hook != nil {
		return func(context *Context) error {
			_, err := hook.Call(nil, context)
			return err
		}
	} else {
		return nil
	}
}

func (self RepresentionFunc) Call(context *Context) {
	// From cache

	cached := false
	if cacheEntry, ok := FromCache(context); ok {
		// Don't use cache entry on GET if there is no body cached
		if context.Context.IsGet() {
			if cacheEntry.Body == nil {
				// Don't use cache entry on GET if there is no body cached
				context.Log.Debugf("cache has no body: %s", context.Path)
			} else {
				cacheEntry.Write(context)
				cached = true
			}
		} else {
			cacheEntry.Write(context)
			cached = true
		}
	}

	if cached {
		if eTag := util.BytesToString(context.Context.Response.Header.Peek(fasthttp.HeaderETag)); eTag != "" {
			ifNoneMatch := util.BytesToString(context.Context.Request.Header.Peek(fasthttp.HeaderIfNoneMatch))
			if ifNoneMatch == context.ETag {
				// The following headers should have been set:
				// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
				context.Context.NotModified()
				return
			}
		}

		if !context.Context.IfModifiedSince(context.LastModified) {
			// The following headers should have been set:
			// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
			context.Context.NotModified()
			return
		}

		return
	}

	// Writer

	if context.Context.IsHead() {
		// Avoid wasting resources on writing for HEAD
		context.Writer = io.Discard
	} else if context.Context.Request.Header.HasAcceptEncoding("gzip") {
		context.Log.Info("gzip!")
		context.Context.Response.Header.Add(fasthttp.HeaderContentEncoding, "gzip")
		context.Writer = NewGZipWriter(context.Writer)
	}

	// Present

	if err := self(context); err != nil {
		context.Log.Errorf("%s", err)
		context.Context.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	context.EndETag()

	if context.LastModified.IsZero() {
		context.LastModified = time.Now()
	}

	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type

	if context.ContentType != "" {
		context.Context.SetContentType(context.ContentType + ";charset=utf-8")
	}

	if context.CacheDuration > 0.0 {

		// Cache-Control
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control

		maxAge := int(context.CacheDuration)
		context.Context.Response.Header.Add(fasthttp.HeaderCacheControl, fmt.Sprintf("max-age=%d", maxAge))

	} else if context.ETag != "" {

		// ETag and If-None-Match
		// (Has precedence over If-Modified-Since)
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match

		var eTag string
		if context.WeakETag {
			eTag = "W/\"" + context.ETag + "\""
		} else {
			eTag = "\"" + context.ETag + "\""
		}

		context.Context.Response.Header.Add(fasthttp.HeaderETag, eTag)

		ifNoneMatch := util.BytesToString(context.Context.Request.Header.Peek(fasthttp.HeaderIfNoneMatch))
		if ifNoneMatch == eTag {
			// The following headers should have been set:
			// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
			context.Context.NotModified()
			return
		}

	} else {
		// Don't store and *also* invalidate the existing client cache
		context.Context.Response.Header.Add(fasthttp.HeaderCacheControl, "no-store,max-age=0")
	}

	// Last-Modified and If-Modified-Since
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since

	if !context.Context.IfModifiedSince(context.LastModified) {
		// The following headers should have been set:
		// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
		context.Context.NotModified()
		return
	}

	if !context.LastModified.IsZero() {
		context.Context.Response.Header.SetLastModified(context.LastModified)
	}

	// GZip

	if gzipWriter, ok := context.Writer.(*GZipWriter); ok {
		if err := gzipWriter.Close(); err != nil {
			context.Log.Errorf("%s", err)
		}
		context.Writer = gzipWriter.Writer
	}

	// To cache

	if context.CacheDuration > 0.0 {
		ToCache(context)
	}
}
