package platform

import (
	"github.com/tliron/kutil/logging"
	urlpkg "github.com/tliron/kutil/url"
)

var log = logging.GetLogger("prudence.platform")

type GetRelativeURL func(url string) (urlpkg.URL, error)

type HasGetRelativeURL interface {
	GetRelativeURL(url string) (urlpkg.URL, error)
}
