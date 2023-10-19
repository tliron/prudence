package rest

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dop251/goja"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
)

//
// RepresentationHook
//

type RepresentationHook func(restContext *Context) error

func GetRepresentationHook(value any, jsContext *commonjs.Context) (RepresentationHook, error) {
	var err error
	if value, jsContext, err = commonjs.Unbind(value, jsContext); err != nil {
		return nil, err
	}

	switch hook := value.(type) {
	case RepresentationHook:
		return hook, nil

	case goja.Value, commonjs.ExportedJavaScriptFunc:
		return func(restContext *Context) error {
			_, err := jsContext.Environment.Call(hook, restContext)
			return err
		}, nil
	}

	return nil, fmt.Errorf("not a representation hook: %T", value)
}

//
// Represention
//

type Representation struct {
	Name                        string
	CharSet                     string
	RedirectTrailingSlash       bool
	RedirectTrailingSlashStatus int
	Variables                   map[string]any
	Prepare                     RepresentationHook
	Describe                    RepresentationHook
	Present                     RepresentationHook
	Erase                       RepresentationHook
	Modify                      RepresentationHook
	Call                        RepresentationHook
}

func NewRepresentation(name string) *Representation {
	return &Representation{
		Name:                        name,
		CharSet:                     "utf-8",
		RedirectTrailingSlashStatus: http.StatusMovedPermanently, // 301
		Variables:                   make(map[string]any),
	}
}

// ([platform.CreateFunc] signature)
func CreateRepresentation(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	name, _ := config_.Get("name").String()

	self := NewRepresentation(name)

	if charSet, ok := config_.Get("charSet").String(); ok {
		self.CharSet = charSet
	}

	self.RedirectTrailingSlash, _ = config_.Get("redirectTrailingSlash").Boolean()

	if redirectTrailingSlashStatus, ok := config_.Get("redirectTrailingSlashStatus").UnsignedInteger(); ok {
		self.RedirectTrailingSlashStatus = int(redirectTrailingSlashStatus)
	}

	if variables, ok := config_.Get("variables").StringMap(); ok {
		self.Variables = variables
	}

	var hooks *ard.Node
	hooksJsContext := jsContext
	if hooks = config_.Get("hooks"); hooks.Value != nil {
		var err error
		if hooks.Value, hooksJsContext, err = commonjs.Unbind(hooks.Value, hooksJsContext); err != nil {
			return nil, err
		}
	}

	getHook := func(name string) (RepresentationHook, error) {
		if hooks.Value != nil {
			if hook := hooks.Get(name).Value; hook != nil {
				return GetRepresentationHook(hook, hooksJsContext)
			}
		}

		if hook := config_.Get(name).Value; hook != nil {
			return GetRepresentationHook(hook, jsContext)
		}

		return nil, nil
	}

	var err error
	if self.Prepare, err = getHook("prepare"); err != nil {
		return nil, err
	}
	if self.Describe, err = getHook("describe"); err != nil {
		return nil, err
	}
	if self.Present, err = getHook("present"); err != nil {
		return nil, err
	}
	if self.Erase, err = getHook("erase"); err != nil {
		return nil, err
	}
	if self.Modify, err = getHook("modify"); err != nil {
		return nil, err
	}
	if self.Call, err = getHook("call"); err != nil {
		return nil, err
	}

	return self, nil
}

// ([Handler] interface, [HandleFunc] signature)
func (self *Representation) Handle(restContext *Context) (bool, error) {
	restContext = restContext.AppendName(self.Name, false)

	ard.Merge(restContext.Variables, self.Variables, false)

	if self.RedirectTrailingSlash {
		if err := restContext.RedirectTrailingSlash(self.RedirectTrailingSlashStatus); err != nil {
			return false, err
		}
	}

	restContext.Response.CharSet = self.CharSet

	switch restContext.Request.Method {
	case "GET":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/GET
		if err := self.prepare(restContext); err == nil {
			if !self.presentFromCache(restContext, true) {
				if ok, err := self.negotiate(restContext); err == nil {
					if ok {
						if err := self.respond(restContext, true); err != nil {
							return false, err
						}
					}
				} else {
					return false, err
				}
			}
		} else {
			return false, err
		}

	case "HEAD":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD

		// Avoid wasting resources on writing
		restContext.Writer = io.Discard

		if err := self.prepare(restContext); err == nil {
			if !self.presentFromCache(restContext, false) {
				if ok, err := self.negotiate(restContext); err == nil {
					if ok {
						if err := self.respond(restContext, false); err != nil {
							return false, err
						}
					}
				} else {
					return false, err
				}
			}
		} else {
			return false, err
		}

	case "DELETE":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/DELETE
		if err := self.prepare(restContext); err == nil {
			if err := self.erase(restContext); err != nil {
				return false, err
			}
		} else {
			return false, err
		}

	case "PUT":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PUT
		if err := self.prepare(restContext); err == nil {
			if err := self.modify(restContext); err != nil {
				return false, err
			}
		} else {
			return false, err
		}

	case "POST":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST
		if err := self.prepare(restContext); err == nil {
			if err := self.call(restContext); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	return restContext.Response.Status != http.StatusNotFound, nil
}

func (self *Representation) prepare(restContext *Context) error {
	//restContext.CacheKey = restContext.Request.Direct.URL.String()

	if self.Prepare != nil {
		return self.Prepare(restContext)
	} else {
		return nil
	}
}

func (self *Representation) presentFromCache(restContext *Context, withBody bool) bool {
	if restContext.CacheKey != "" {
		if key, cached, ok := restContext.LoadCachedRepresentation(); ok {
			if withBody && (len(cached.Body) == 0) {
				// The cache entry was likely created by a previous HEAD request
				restContext.Log.Debugf("ignoring cached representation because it has no body: %s", restContext.Request.Path)
			} else {
				if changed := restContext.PresentCachedRepresentation(cached, withBody); changed {
					restContext.UpdateCachedRepresentation(key, cached)
				}
				return true
			}
		}
	}

	return false
}

func (self *Representation) negotiate(restContext *Context) (bool, error) {
	if self.Describe != nil {
		if err := self.Describe(restContext); err == nil {
			return !restContext.isNotModified(false), nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func (self *Representation) respond(restContext *Context, withBody bool) error {
	if withBody {
		// Encoding
		if !SetBestEncodeWriter(restContext) {
			return nil
		}

		// Present
		if self.Present != nil {
			if err := self.Present(restContext); err != nil {
				return err
			}
		}

		if restContext.isNotModified(false) {
			return nil
		}
	}

	if err := restContext.Flush(); err != nil {
		return err
	}

	// Headers
	restContext.Response.setContentType()
	restContext.Response.setETag()
	restContext.Response.setLastModified()
	restContext.setCacheControl()

	if restContext.caching() {
		restContext.StoreCachedRepresentation(withBody)
	}

	return nil
}

func (self *Representation) erase(restContext *Context) error {
	if self.Erase != nil {
		if err := self.Erase(restContext); err != nil {
			return err
		}

		if restContext.Done {
			if restContext.Async {
				// Will be erased later
				restContext.Response.Status = http.StatusAccepted // 202
			} else if restContext.Response.Buffer.Len() > 0 {
				// Erased, has response
				restContext.Response.Status = http.StatusOK // 200
			} else {
				// Erased, no response
				restContext.Response.Status = http.StatusNoContent // 204
			}

			if restContext.CacheKey != "" {
				restContext.DeleteCachedRepresentation()
			}
		} else {
			restContext.Response.Status = http.StatusNotFound // 404
		}
	} else {
		restContext.Response.Status = http.StatusMethodNotAllowed // 405
	}

	return nil
}

func (self *Representation) modify(restContext *Context) error {
	if self.Modify != nil {
		if err := self.Modify(restContext); err != nil {
			return err
		}

		if restContext.Done {
			if restContext.Created {
				// Created
				restContext.Response.Status = http.StatusCreated // 201
			} else if restContext.Response.Buffer.Len() > 0 {
				// Changed, has response
				restContext.Response.Status = http.StatusOK // 200
			} else {
				// Changed, no response
				restContext.Response.Status = http.StatusNoContent // 204
			}

			if restContext.caching() {
				restContext.StoreCachedRepresentation(true)
			}
		} else {
			restContext.Response.Status = http.StatusNotFound // 404
		}
	} else {
		restContext.Response.Status = http.StatusMethodNotAllowed // 405
	}

	return nil
}

func (self *Representation) call(restContext *Context) error {
	if self.Call != nil {
		return self.Call(restContext)
	} else {
		return nil
	}
}
