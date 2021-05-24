package rest

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dop251/goja"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/platform"
)

//
// RepresentionFunc
//

type RepresentionFunc func(context *Context) error

func NewRepresentationFunc(function interface{}, runtime *goja.Runtime) (RepresentionFunc, error) {
	if call, ok := function.(func(goja.FunctionCall) goja.Value); ok {
		return func(context *Context) error {
			call(goja.FunctionCall{
				This:      nil,
				Arguments: []goja.Value{runtime.ToValue(context)},
			})
			// TODO: exceptions??
			return nil
		}, nil
	} else {
		return nil, fmt.Errorf("not a function: %T", function)
	}
}

//
// Represention
//

type Representation struct {
	Construct RepresentionFunc
	Describe  RepresentionFunc
	Present   RepresentionFunc
	Erase     RepresentionFunc
	Change    RepresentionFunc
}

func CreateRepresentation(node *ard.Node) (*Representation, error) {
	var self Representation

	var get func(name string) (RepresentionFunc, error)

	if functions, ok := node.Get("functions").StringMap(false); ok {
		// "functions" property
		if runtime, ok := functions["runtime"]; ok {
			if runtime_, ok := runtime.(*goja.Runtime); ok {
				get = func(name string) (RepresentionFunc, error) {
					if f, ok := functions[name]; ok {
						return NewRepresentationFunc(f, runtime_)
					} else {
						return nil, nil
					}
				}
			} else {
				return nil, errors.New("invalid \"runtime\" property in \"functions\"")
			}
		} else {
			return nil, errors.New("no \"runtime\" property in \"functions\"")
		}
	} else {
		// Individual function properties
		get = func(name string) (RepresentionFunc, error) {
			if f := node.Get(name).Data; f != nil {
				if runtime, ok := node.Get("runtime").Data.(*goja.Runtime); ok {
					return NewRepresentationFunc(f, runtime)
				} else {
					return nil, errors.New("no valid \"runtime\" property")
				}
			} else {
				return nil, nil
			}
		}
	}

	var err error
	if self.Construct, err = get("construct"); err != nil {
		return nil, err
	}
	if self.Describe, err = get("describe"); err != nil {
		return nil, err
	}
	if self.Present, err = get("present"); err != nil {
		return nil, err
	}
	if self.Erase, err = get("erase"); err != nil {
		return nil, err
	}
	if self.Change, err = get("change"); err != nil {
		return nil, err
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

	if context.context.IsDelete() {
		// Erase
		if self.Erase != nil {
			if err := self.Erase(context); err != nil {
				context.Error(err)
				return true
			}
		}
	} else if context.context.IsPut() {
		// Change
		if self.Change != nil {
			if err := self.Change(context); err != nil {
				context.Error(err)
				return true
			}
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
