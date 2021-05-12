package rest

import (
	"fmt"
	"io"
	"time"

	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

//
// RepresentFunc
//

type RepresentFunc func(context *Context)

// Handler interface
// HandleFunc signature
func (self RepresentFunc) Handle(context *Context) bool {
	// From cache

	if cacheEntry, ok := FromCache(context); ok {
		// Don't use cache entry on GET if there is no body cached
		if context.Context.IsGet() {
			if cacheEntry.Body == nil {
				// Don't use cache entry on GET if there is no body cached
				context.Log.Debugf("cache has no body: %s", context.Path)
			} else {
				cacheEntry.Write(context)
				return true
			}
		} else {
			cacheEntry.Write(context)
			return true
		}
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

	// Represent
	self(context)

	if context.LastModified.IsZero() {
		context.LastModified = time.Now()
	}

	// ETag

	if hashWriter, ok := context.Writer.(*HashWriter); ok {
		// Form AutoETag
		context.ETag = hashWriter.Hash()
	}
	if context.ETag != "" {
		if context.WeakETag {
			context.ETag = "W/\"" + context.ETag + "\""
		} else {
			context.ETag = "\"" + context.ETag + "\""
		}
	}

	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type

	context.Context.Response.Header.SetNoDefaultContentType(true)
	if context.ContentType != "" {
		context.Context.SetContentType(context.ContentType)
	}

	// Cache-Control
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control

	if context.CacheDuration > 0.0 {
		maxAge := int(context.CacheDuration)
		context.Context.Response.Header.Add(fasthttp.HeaderCacheControl, fmt.Sprintf("max-age=%d", maxAge))
	} else {
		// Don't store and *also* invalid the existing client cache
		context.Context.Response.Header.Add(fasthttp.HeaderCacheControl, "no-store,max-age=0")
	}

	// ETag and If-None-Match
	// (Has precedence over If-Modified-Since)
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match

	if context.ETag != "" {
		context.Context.Response.Header.Add(fasthttp.HeaderETag, context.ETag)

		ifNoneMatch := util.BytesToString(context.Context.Request.Header.Peek(fasthttp.HeaderIfNoneMatch))
		if ifNoneMatch == context.ETag {
			// The following headers should have been set:
			// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
			context.Context.NotModified()
			return true
		}
	}

	// Last-Modified and If-Modified-Since
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since

	context.Context.Response.Header.SetLastModified(context.LastModified)

	if !context.Context.IfModifiedSince(context.LastModified) {
		context.Context.NotModified()
		return true
	}

	// Writer

	if eTagBuffer, ok := context.Writer.(*HashWriter); ok {
		eTagBuffer.Close()
		context.Writer = eTagBuffer.Writer
	}

	if gzipWriter, ok := context.Writer.(*GZipWriter); ok {
		context.Writer = gzipWriter.Writer
	}

	// To cache

	if context.CacheDuration > 0.0 {
		ToCache(context)
	}

	return true
}
