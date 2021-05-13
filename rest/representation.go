package rest

import (
	"fmt"
	"io"
	"time"

	"github.com/tliron/kutil/js"
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

	if cacheEntry, ok := FromCache(context); ok {
		// Don't use cache entry on GET if there is no body cached
		if context.context.IsGet() {
			if cacheEntry.Body == nil {
				// Don't use cache entry on GET if there is no body cached
				context.Log.Debugf("cache has no body: %s", context.Path)
			} else {
				cacheEntry.Write(context)
				return
			}
		} else {
			cacheEntry.Write(context)
			return
		}
	}

	// Writer

	if context.context.IsHead() {
		// Avoid wasting resources on writing for HEAD
		context.Writer = io.Discard
	} else if context.context.Request.Header.HasAcceptEncoding("gzip") {
		context.Log.Info("gzip!")
		context.context.Response.Header.Add(fasthttp.HeaderContentEncoding, "gzip")
		context.Writer = NewGZipWriter(context.Writer)
	}

	// Present

	if err := self(context); err != nil {
		context.Log.Errorf("%s", err)
		context.context.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	if context.LastModified.IsZero() {
		context.LastModified = time.Now()
	}

	context.EndETag()

	var eTag string
	if context.ETag != "" {
		if context.WeakETag {
			eTag = "W/\"" + context.ETag + "\""
		} else {
			eTag = "\"" + context.ETag + "\""
		}
	}

	if context.CacheDuration > 0.0 {

		// Enabling caching means no conditional checks

		// Cache-Control
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
		maxAge := int(context.CacheDuration)
		AddCacheControl(context.context, fmt.Sprintf("max-age=%d", maxAge))

	} else {

		// Conditional

		// If-None-Match
		// (Has precedence over If-Modified-Since)
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
		if IfNoneMatch(context.context, eTag) {
			context.context.NotModified()
			return
		}

		// If-Modified-Since
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
		if !context.context.IfModifiedSince(context.LastModified) {
			context.context.NotModified()
			return
		}

	}

	// Last-Modified
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	context.context.Response.Header.SetLastModified(context.LastModified)

	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
	if context.ContentType != "" {
		context.context.SetContentType(context.ContentType + ";charset=utf-8")
	}

	// ETag
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	if eTag != "" {
		context.context.Response.Header.Add(fasthttp.HeaderETag, eTag)
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
