package common

import (
	urlpkg "github.com/tliron/kutil/url"
)

type GetRelativeURL func(url string) (urlpkg.URL, error)
