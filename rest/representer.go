package rest

import (
	"fmt"
	"io"

	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

//
// Representer
//

type Representer func(context *Context)

// Handler interface
// HandlerFunc signature
func (self Representer) Handle(context *Context) bool {
	if context.RequestContext.IsHead() {
		// Avoid wasting resources on writing
		context.Writer = io.Discard
	} else if context.RequestContext.Request.Header.HasAcceptEncoding("gzip") {
		context.Log.Info("gzip!")
		context.RequestContext.Response.Header.Add(fasthttp.HeaderContentEncoding, "gzip")
		context.Writer = NewGZipWriter(context.Writer)
	}

	self(context)

	if eTagBuffer, ok := context.Writer.(*ETagBuffer); ok {
		// AutoETag
		context.ETag = eTagBuffer.ETag()
	}

	if context.ETag != "" {
		// Strong ETag
		context.ETag = "\"" + context.ETag + "\""
	}

	if context.ContentType != "" {
		context.RequestContext.SetContentType(context.ContentType)
	}

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	if context.MaxAge >= 0 {
		context.RequestContext.Response.Header.Add(fasthttp.HeaderCacheControl, fmt.Sprintf("max-age=%d", context.MaxAge))
	}

	if context.ETag != "" {
		context.RequestContext.Response.Header.DisableNormalizing() // so we don't change ETag -> Etag
		context.RequestContext.Response.Header.Add(fasthttp.HeaderETag, context.ETag)

		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
		// Has precedence over If-Modified-Since
		ifNoneMatch := util.BytesToString(context.RequestContext.Request.Header.Peek(fasthttp.HeaderIfNoneMatch))
		if ifNoneMatch == context.ETag {
			context.RequestContext.NotModified()
			// The following headers should have been set:
			// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
			return true
		}
	}

	if !context.LastModified.IsZero() {
		context.RequestContext.Response.Header.SetLastModified(context.LastModified)

		if context.RequestContext.IfModifiedSince(context.LastModified) {
			context.RequestContext.NotModified()
			return true
		}
	}

	if eTagBuffer, ok := context.Writer.(*ETagBuffer); ok {
		eTagBuffer.Close()
		context.Writer = eTagBuffer.Writer
	}

	if gzipWriter, ok := context.Writer.(*GZipWriter); ok {
		context.Writer = gzipWriter.Writer
	}

	return true
}
