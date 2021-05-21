package rest

import (
	"fmt"
	"io"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
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

//
// Represention
//

type Representation struct {
	Construct RepresentionFunc
	Describe  RepresentionFunc
	Present   RepresentionFunc
}

func CreateRepresentation(node *ard.Node) (*Representation, error) {
	var self Representation

	if hooks := node.Get("hooks").Data; hooks != nil {
		hooks_ := hooks.(map[string]*js.Hook)
		if construct, ok := hooks_["construct"]; ok {
			self.Construct = NewRepresentationFunc(construct)
		}
		if describe, ok := hooks_["describe"]; ok {
			self.Describe = NewRepresentationFunc(describe)
		}
		if present, ok := hooks_["present"]; ok {
			self.Present = NewRepresentationFunc(present)
		}
	}

	if construct := node.Get("construct").Data; construct != nil {
		self.Construct = NewRepresentationFunc(construct.(*js.Hook))
	}
	if describe := node.Get("describe").Data; describe != nil {
		self.Describe = NewRepresentationFunc(describe.(*js.Hook))
	}
	if present := node.Get("present").Data; present != nil {
		self.Present = NewRepresentationFunc(present.(*js.Hook))
	}

	return &self, nil
}

// Handler interface
// HandleFunc signature
func (self *Representation) Handle(context *Context) bool {
	context.CacheKey = context.context.URI().String()
	context.CharSet = "utf-8"

	// Construct
	if self.Construct != nil {
		if err := self.Construct(context); err != nil {
			context.Error(err)
			return true
		}
	}

	// Try cache
	if context.CacheKey != "" {
		if cacheKey, cacheEntry, ok := CacheLoad(context); ok {
			if context.context.IsHead() {
				// HEAD doesn't care if the cacheEntry doesn't have a body
				if changed := CacheEntryToContext(cacheEntry, context); changed {
					CacheUpdate(cacheKey, cacheEntry)
				}
				return !NotFound(context.context)
			} else {
				if len(cacheEntry.Body) == 0 {
					context.Log.Debugf("ignoring cache with no body: %s", context.Path)
				} else {
					if changed := CacheEntryToContext(cacheEntry, context); changed {
						CacheUpdate(cacheKey, cacheEntry)
					}
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
	if self.Describe != nil {
		if err := self.Describe(context); err != nil {
			context.Error(err)
			return true
		}
	}

	if !context.context.IsHead() {
		// Encoding
		SetBestEncodeWriter(context)

		// Present
		if self.Present != nil {
			if err := self.Present(context); err != nil {
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

	// Unwrap writers
	for true {
		if wrappedWriter, ok := context.writer.(WrappingWriter); ok {
			if err := wrappedWriter.Close(); err != nil {
				context.Error(err)
			}
			context.writer = wrappedWriter.GetWrappedWriter()
		} else {
			break
		}
	}

	// To cache
	if (context.CacheDuration > 0.0) && (context.CacheKey != "") {
		CacheStoreContext(context)
	}

	return !NotFound(context.context)
}

//
// Representations
//

type Representations map[string]*Representation

func CreateRepresentations(config ard.Value) (Representations, error) {
	self := make(Representations)

	representations := platform.AsConfigList(config)
	for _, representation := range representations {
		representation_ := ard.NewNode(representation)
		representation__, _ := CreateRepresentation(representation_)
		contentTypes := platform.AsStringList(representation_.Get("contentTypes").Data)
		// TODO:
		//charSets := asStringList(representation_.Get("charSets").Data)
		//languages := asStringList(representation_.Get("languages").Data)

		if len(contentTypes) == 0 {
			// Default representation
			self[""] = representation__
		} else {
			for _, contentType := range contentTypes {
				self[contentType] = representation__
			}
		}
	}

	return self, nil
}
