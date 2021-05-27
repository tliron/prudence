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
	Call      RepresentionFunc
}

func CreateRepresentation(node *ard.Node) (*Representation, error) {
	var self Representation

	var get func(name string) (RepresentionFunc, error)

	if functions := node.Get("functions"); functions.Data != nil {
		var runtime *goja.Runtime
		var ok bool
		if runtime, ok = functions.Get("runtime").Data.(*goja.Runtime); !ok {
			if runtime, ok = node.Get("runtime").Data.(*goja.Runtime); !ok {
				return nil, errors.New("no valid \"runtime\" property")
			}
		}

		get = func(name string) (RepresentionFunc, error) {
			if f := functions.Get(name).Data; f != nil {
				return NewRepresentationFunc(f, runtime)
			} else {
				return nil, nil
			}
		}
	} else {
		var runtime *goja.Runtime
		var ok bool
		if runtime, ok = node.Get("runtime").Data.(*goja.Runtime); !ok {
			return nil, errors.New("no valid \"runtime\" property")
		}

		// Individual function properties
		get = func(name string) (RepresentionFunc, error) {
			if f := node.Get(name).Data; f != nil {
				return NewRepresentationFunc(f, runtime)
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
	if self.Call, err = get("call"); err != nil {
		return nil, err
	}

	return &self, nil
}

// Handler interface
// HandleFunc signature
func (self *Representation) Handle(context *Context) bool {
	context.CharSet = "utf-8"

	switch context.Method {
	case "GET":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/GET
		if self.construct(context) {
			if self.tryCache(context, true) {
				if self.describe(context) {
					self.present(context, true)
				}
			}
		}

	case "HEAD":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD

		// Avoid wasting resources on writing for HEAD
		context.writer = io.Discard

		if self.construct(context) {
			if self.tryCache(context, false) {
				if self.describe(context) {
					self.present(context, false)
				}
			}
		}

	case "DELETE":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/DELETE
		if self.construct(context) {
			self.erase(context)
		}

	case "PUT":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PUT
		if self.construct(context) {
			self.change(context)
		}

	case "POST":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST
		if self.construct(context) {
			self.call(context)
		}
	}

	return !NotFound(context.Context)
}

func (self *Representation) construct(context *Context) bool {
	context.CacheKey = context.Context.URI().String()
	if self.Construct != nil {
		if err := self.Construct(context); err != nil {
			context.Error(err)
			return false
		}
	}

	return true
}

func (self *Representation) tryCache(context *Context, withBody bool) bool {
	if context.CacheKey != "" {
		if cacheKey, cacheEntry, ok := CacheLoad(context); ok {
			if withBody && (len(cacheEntry.Body) == 0) {
				// The cache entry was likely created by a previous HEAD request
				context.Log.Debugf("ignoring cache becase it has no body: %s", context.Path)
			} else {
				if changed := CacheEntryPresent(cacheEntry, context, withBody); changed {
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

func (self *Representation) present(context *Context, withBody bool) {
	if withBody {
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
		CacheStoreContext(context, withBody)
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
				CacheStoreContext(context, true)
			}
		} else {
			context.Context.SetStatusCode(fasthttp.StatusNotFound) // 404
		}
	} else {
		context.Context.SetStatusCode(fasthttp.StatusMethodNotAllowed) // 405
	}
}

func (self *Representation) call(context *Context) {
	if self.Call != nil {
		if err := self.Call(context); err != nil {
			context.Error(err)
			return
		}
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
