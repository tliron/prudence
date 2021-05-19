package js

import (
	"github.com/tliron/kutil/logging"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/prudence/rest"

	_ "github.com/tliron/prudence/jst"
	_ "github.com/tliron/prudence/memory"
	_ "github.com/tliron/prudence/render"
)

//
// PrudenceAPI
//

type PrudenceAPI struct {
	Url             urlpkg.URL
	Log             logging.Logger
	DefaultNotFound rest.Handler
}

func NewPrudenceAPI(url urlpkg.URL) *PrudenceAPI {
	return &PrudenceAPI{
		Url:             url,
		Log:             log,
		DefaultNotFound: rest.DefaultNotFound,
	}
}
