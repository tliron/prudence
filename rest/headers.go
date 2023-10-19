package rest

import (
	"net/http"
	timepkg "time"
)

func GetTimeHeader(key string, header http.Header) (timepkg.Time, bool) {
	if lastModified := header.Get(key); lastModified != "" {
		if time, err := http.ParseTime(lastModified); err == nil {
			return time, true
		}
	}

	return timepkg.Time{}, false
}

func SetLastModifiedHeader(header http.Header, time timepkg.Time) {
	time = time.UTC()
	header.Set(HeaderLastModified, time.Format(http.TimeFormat))
}

func SetContentTypeHeader(header http.Header, contentType string, charSet string) {
	if charSet != "" {
		contentType += "; charset=" + charSet
	}
	header.Set(HeaderContentType, contentType)
}
