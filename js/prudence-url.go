package js

import (
	urlpkg "github.com/tliron/kutil/url"
)

// common.HasGetRelativeURL interface
// common.GetRelativeURL signature
func (self *PrudenceAPI) GetRelativeURL(url string) (urlpkg.URL, error) {
	urlContext := urlpkg.NewContext()
	defer urlContext.Release()

	var origins []urlpkg.URL
	if self.Url != nil {
		origins = []urlpkg.URL{self.Url.Origin()}
	}

	return urlpkg.NewValidURL(url, origins, urlContext)
}

func (self *PrudenceAPI) Load(url string) (string, error) {
	if url_, err := self.GetRelativeURL(url); err == nil {
		return urlpkg.ReadString(url_)
	} else {
		return "", err
	}
}

func (self *PrudenceAPI) Download(sourceUrl string, targetPath string) error {
	if sourceUrl_, err := self.GetRelativeURL(sourceUrl); err == nil {
		return urlpkg.DownloadTo(sourceUrl_, targetPath)
	} else {
		return err
	}
}
