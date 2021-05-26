package rest

import (
	"errors"
	"fmt"
	"io"

	"github.com/dop251/goja"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

//
// RepresentionFunc
//

type RepresentionFunc func(context *Context) error

func NewRepresentationFunc(function interface{}, runtime *goja.Runtime) (RepresentionFunc, error) {
	if function_, ok := function.(JavaScriptFunc); ok {
		return func(context *Context) error {
			CallJavaScript(runtime, function_, context)
			return nil
		}, nil
	} else {
		return nil, fmt.Errorf("not a JavaScript function: %T", function)
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
	context.CharSet = "utf-8"

	if context.Context.IsHead() {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD

		// Avoid wasting resources on writing for HEAD
		context.writer = io.Discard

		if self.construct(context, true) {
			if self.describe(context) {
				self.present(context)
			}
		}
	} else if context.Context.IsGet() {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/GET
		if self.construct(context, true) {
			if self.describe(context) {
				self.present(context)
			}
		}
	} else if context.Context.IsDelete() {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/DELETE
		if self.construct(context, false) {
			self.erase(context)
		}
	} else if context.Context.IsPut() {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PUT
		if self.construct(context, false) {
			self.change(context)
		}
	} else {
		// TODO
	}

	return !NotFound(context.Context)
}

func (self *Representation) construct(context *Context, tryCache bool) bool {
	context.CacheKey = context.Context.URI().String()
	if self.Construct != nil {
		if err := self.Construct(context); err != nil {
			context.Error(err)
			return false
		}
	}

	// Try cache
	if tryCache && (context.CacheKey != "") {
		if cacheKey, cacheEntry, ok := CacheLoad(context); ok {
			if context.Context.IsGet() && (len(cacheEntry.Body) == 0) {
				// The cache entry was likely created by a previous HEAD request
				context.Log.Debugf("ignoring cache becase it has no body: %s", context.Path)
			} else {
				if changed := CacheEntryPresent(cacheEntry, context); changed {
					CacheUpdate(cacheKey, cacheEntry)
				}
				return false
			}
		}
	}

	return true
}

func (self *Representation) describe(context *Context) bool {
	if self.Describe != nil {
		if err := self.Describe(context); err != nil {
			context.Error(err)
			return false
		}

		if context.isNotModified() {
			return false
		}
	}

	return true
}

func (self *Representation) present(context *Context) {
	if context.Context.IsGet() {
		// Encoding
		SetBestEncodeWriter(context)

		// Present
		if self.Present != nil {
			if err := self.Present(context); err != nil {
				context.Error(err)
				return
			}
		}

		if context.isNotModified() {
			return
		}
	}

	context.unwrapWriters()

	// Headers
	context.setContentType()
	context.addETag()
	context.setLastModified()
	context.setCacheControl()

	if (context.CacheDuration > 0.0) && (context.CacheKey != "") {
		CacheStoreContext(context)
	}
}

func (self *Representation) erase(context *Context) {
	if self.Erase != nil {
		if err := self.Erase(context); err != nil {
			context.Error(err)
			return
		}

		if context.Done {
			if context.Async {
				// Will be erased later
				context.Context.SetStatusCode(fasthttp.StatusAccepted) // 202
			} else if len(context.Context.Response.Body()) > 0 {
				// Erased, has response
				context.Context.SetStatusCode(fasthttp.StatusOK) // 200
			} else {
				// Erased, no response
				context.Context.SetStatusCode(fasthttp.StatusNoContent) // 204
			}

			if context.CacheKey != "" {
				CacheDelete(context)
			}
		} else {
			context.Context.SetStatusCode(fasthttp.StatusNotFound) // 404
		}
	} else {
		context.Context.SetStatusCode(fasthttp.StatusMethodNotAllowed) // 405
	}
}

func (self *Representation) change(context *Context) {
	if self.Change != nil {
		if err := self.Change(context); err != nil {
			context.Error(err)
			return
		}

		if context.Done {
			if context.Created {
				// Created
				context.Context.SetStatusCode(fasthttp.StatusCreated) // 201
			} else if len(context.Context.Response.Body()) > 0 {
				// Erased, has response
				context.Context.SetStatusCode(fasthttp.StatusOK) // 200
			} else {
				// Erased, no response
				context.Context.SetStatusCode(fasthttp.StatusNoContent) // 204
			}

			if (context.CacheDuration > 0.0) && (context.CacheKey != "") {
				CacheStoreContext(context)
			}
		} else {
			context.Context.SetStatusCode(fasthttp.StatusNotFound) // 404
		}
	} else {
		context.Context.SetStatusCode(fasthttp.StatusMethodNotAllowed) // 405
	}
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
		if representation__, err := CreateRepresentation(representation_); err == nil {
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
		} else {
			return nil, err
		}
	}

	return self, nil
}
