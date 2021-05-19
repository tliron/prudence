package rest

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterCreator("facet", CreateFacet)
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
func CreateFacet(config ard.StringMap, getRelativeURL platform.GetRelativeURL) (interface{}, error) {
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
	self.Representations, _ = CreateRepresentations(config_.Get("representations").Data)

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
	var representation *Representation
	var ok bool
	if representation, context.ContentType, ok = self.FindRepresentation(context); !ok {
		return false
	}

	context = context.Copy()
	context.CacheKey = context.context.URI().String()
	context.CharSet = "utf-8"

	// Construct
	if representation.Construct != nil {
		if err := representation.Construct(context); err != nil {
			context.Error(err)
			return true
		}
	}

	// Try cache
	if context.CacheKey != "" {
		if cacheEntry, ok := CacheLoad(context); ok {
			if context.context.IsHead() {
				// HEAD doesn't care if the cacheEntry doesn't have a body
				CacheEntryToContext(cacheEntry, context)
				return !NotFound(context.context)
			} else {
				if len(cacheEntry.Body) == 0 {
					context.Log.Debugf("ignoring cache with no body: %s", context.Path)
				} else {
					CacheEntryToContext(cacheEntry, context)
					return !NotFound(context.context)
				}
			}
		}
	}

	if context.context.IsHead() {
		// Avoid wasting resources on writing for HEAD
		context.writer = io.Discard
	}

	// Describe
	if representation.Describe != nil {
		if err := representation.Describe(context); err != nil {
			context.Error(err)
			return true
		}
	}

	if !context.context.IsHead() {
		// Encoding
		SetBestEncodeWriter(context)

		// Present
		if representation.Present != nil {
			if err := representation.Present(context); err != nil {
				context.Error(err)
				return true
			}
		}
	}

	if context.LastModified.IsZero() {
		context.LastModified = time.Now()
	}

	context.EndSignature()

	eTag := context.ETag()

	if context.CacheDuration < 0.0 {

		// Don't store and *also* invalidate the existing client cache
		AddCacheControl(context.context, "no-store,max-age=0")

	} else if context.CacheDuration > 0.0 {

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

	// Content-Type
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
	if context.ContentType != "" {
		if context.CharSet != "" {
			context.context.SetContentType(context.ContentType + ";charset=" + context.CharSet)
		} else {
			context.context.SetContentType(context.ContentType)
		}
	}

	// Last-Modified
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	context.context.Response.Header.SetLastModified(context.LastModified)

	// ETag
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	if eTag != "" {
		AddETag(context.context, eTag)
	}

	// Encode
	if encodeWriter, ok := context.writer.(*EncodeWriter); ok {
		if err := encodeWriter.Close(); err != nil {
			context.Error(err)
		}
		context.writer = encodeWriter.Writer
	}

	// To cache
	if (context.CacheDuration > 0.0) && (context.CacheKey != "") {
		CacheStoreContext(context)
	}

	return !NotFound(context.context)
}
