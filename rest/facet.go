package rest

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/js/common"
	"github.com/valyala/fasthttp"
)

func init() {
	Register("facet", CreateFacet)
}

//
// Facet
//

type Facet struct {
	*Route

	Representations Representations
}

func NewFacet(name string, paths []string) *Facet {
	self := Facet{
		Route:           NewRoute(name, paths, nil),
		Representations: make(Representations),
	}
	self.Handler = self.Handle
	return &self
}

// CreateFunc signature
func CreateFacet(config ard.StringMap, getRelativeURL common.GetRelativeURL) (interface{}, error) {
	self := Facet{
		Representations: make(Representations),
	}

	route, _ := CreateRoute(config, getRelativeURL)
	self.Route = route.(*Route)
	if self.Handler != nil {
		return nil, errors.New("cannot set \"handler\" on facet")
	}
	self.Handler = self.Handle

	config_ := ard.NewNode(config)
	representations, _ := config_.Get("representations").List(true)
	self.Representations, _ = CreateRepresentations(representations)

	return &self, nil
}

func (self *Facet) FindRepresentation(context *Context) (*Representation, string, bool) {
	for _, contentType := range ParseAccept(context) {
		if functions, ok := self.Representations[contentType]; ok {
			return functions, contentType, true
		}
	}

	// Default representation
	functions, ok := self.Representations[""]
	return functions, "", ok
}

// Handler interface
// HandleFunc signature
func (self *Facet) Handle(context *Context) bool {
	context = context.Copy()

	var representation *Representation
	var ok bool
	if representation, context.ContentType, ok = self.FindRepresentation(context); !ok {
		return false
	}

	context.CacheKey = context.context.URI().String()

	if representation.Construct != nil {
		if err := representation.Construct(context); err != nil {
			context.Log.Errorf("%s", err)
			context.context.SetStatusCode(fasthttp.StatusInternalServerError)
			return true
		}
	}

	// Try cache

	if cacheEntry, ok := FromCache(context); ok {
		if context.context.IsHead() {
			// HEAD doesn't care if the cacheEntry doesn't have a body
			cacheEntry.Write(context)
			return !NotFound(context.context)
		} else {
			if cacheEntry.Body == nil {
				context.Log.Debugf("ignoring cache with no body: %s", context.Path)
			} else {
				cacheEntry.Write(context)
				return !NotFound(context.context)
			}
		}
	}

	if context.context.IsHead() {
		// Avoid wasting resources on writing for HEAD
		context.writer = io.Discard
	}

	if representation.Describe != nil {
		if err := representation.Describe(context); err != nil {
			context.Log.Errorf("%s", err)
			context.context.SetStatusCode(fasthttp.StatusInternalServerError)
			return true
		}
	}

	if !context.context.IsHead() {
		if context.context.Request.Header.HasAcceptEncoding("gzip") {
			context.Log.Info("gzip!")
			AddContentEncoding(context.context, "gzip")
			context.writer = NewGZipWriter(context.writer)
		}

		if representation.Present != nil {
			if err := representation.Present(context); err != nil {
				context.Log.Errorf("%s", err)
				context.context.SetStatusCode(fasthttp.StatusInternalServerError)
				return true
			}
		}
	}

	if context.LastModified.IsZero() {
		context.LastModified = time.Now()
	}

	context.EndETag()

	eTag := context.RenderETag()

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
			return true
		}

		// If-Modified-Since
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
		if !context.context.IfModifiedSince(context.LastModified) {
			context.context.NotModified()
			return true
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
		AddETag(context.context, eTag)
	}

	// GZip

	if gzipWriter, ok := context.writer.(*GZipWriter); ok {
		if err := gzipWriter.Close(); err != nil {
			context.Log.Errorf("%s", err)
		}
		context.writer = gzipWriter.Writer
	}

	// To cache

	if context.CacheDuration > 0.0 {
		ToCache(context)
	}

	return !NotFound(context.context)
}
