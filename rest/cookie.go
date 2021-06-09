package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/js"
	"github.com/tliron/prudence/platform"
)

func init() {
	platform.RegisterType("Cookie", CreateCookie)
}

// CreateFunc signature
func CreateCookie(config ard.StringMap, context *js.Context) (interface{}, error) {
	var self http.Cookie

	config_ := ard.NewNode(config)
	var ok bool
	if self.Name, ok = config_.Get("name").String(false); !ok {
		return nil, errors.New("must set cookie \"name\"")
	}
	if self.Value, ok = config_.Get("value").String(false); !ok {
		return nil, errors.New("must set cookie \"value\"")
	}
	self.Path, _ = config_.Get("path").String(true)
	self.Domain, _ = config_.Get("domain").String(true)
	if expires := config_.Get("expires").Data; expires != nil {
		if self.Expires, ok = expires.(time.Time); !ok {
			return nil, fmt.Errorf("invalid cookie \"expires\": %T", expires)
		}
	}
	if maxAge, ok := config_.Get("maxAge").Integer(false); ok {
		self.MaxAge = int(maxAge)
	}
	self.Secure, _ = config_.Get("secure").Boolean(true)
	self.HttpOnly, _ = config_.Get("httpOnly").Boolean(true)
	if sameSite, ok := config_.Get("sameSite").String(false); ok {
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
			return nil, fmt.Errorf("invalid cookie \"sameSite\": %s", sameSite)
		}
	}

	return &self, nil
}
