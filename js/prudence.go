package js

import (
	"github.com/tliron/kutil/js"
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
	js.UtilAPI
	js.FormatAPI
	js.FileAPI
	js.URLAPI

	Log             logging.Logger
	DefaultNotFound rest.Handler
}

func NewPrudenceAPI(url urlpkg.URL) *PrudenceAPI {
	return &PrudenceAPI{
		URLAPI:          js.URLAPI{Url: url},
		Log:             log,
		DefaultNotFound: rest.DefaultNotFound,
	}
}
