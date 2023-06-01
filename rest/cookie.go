package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Cookie", CreateCookie)
}

// CreateFunc signature
func CreateCookie(config ard.StringMap, context *commonjs.Context) (interface{}, error) {
	var self http.Cookie

	config_ := ard.NewNode(config)
	var ok bool
	if self.Name, ok = config_.Get("name").String(); !ok {
		return nil, errors.New("Cookie must have a \"name\"")
	}
	if self.Value, ok = config_.Get("value").String(); !ok {
		return nil, errors.New("Cookie must have a \"value\"")
	}
	self.Path, _ = config_.Get("path").String()
	self.Domain, _ = config_.Get("domain").String()
	if expires := config_.Get("expires").Value; expires != nil {
		if self.Expires, ok = expires.(time.Time); !ok {
			return nil, fmt.Errorf("invalid cookie \"expires\": %T", expires)
		}
	}
	if maxAge, ok := config_.Get("maxAge").Integer(); ok {
		self.MaxAge = int(maxAge)
	}
	self.Secure, _ = config_.Get("secure").Boolean()
	self.HttpOnly, _ = config_.Get("httpOnly").Boolean()
	if sameSite, ok := config_.Get("sameSite").String(); ok {
		switch sameSite {
		case "default":
			self.SameSite = http.SameSiteDefaultMode
		case "lax":
			self.SameSite = http.SameSiteLaxMode
		case "strict":
			self.SameSite = http.SameSiteStrictMode
		case "none":
			self.SameSite = http.SameSiteNoneMode
		default:
			return nil, fmt.Errorf("Cookie has invalid \"sameSite\": %s", sameSite)
		}
	}

	return &self, nil
}
