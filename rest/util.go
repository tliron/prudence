package rest

import (
	"strings"
	"time"

	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

func GetETag(context *fasthttp.RequestCtx) string {
	return util.BytesToString(context.Response.Header.Peek(fasthttp.HeaderETag))
}

func AddETag(context *fasthttp.RequestCtx, eTag string) {
	context.Response.Header.Add(fasthttp.HeaderETag, eTag)
}

func GetContentEncoding(context *fasthttp.RequestCtx) string {
	return util.BytesToString(context.Response.Header.Peek(fasthttp.HeaderContentEncoding))
}

func GetLastModified(context *fasthttp.RequestCtx) time.Time {
	if lastModified, err := fasthttp.ParseHTTPDate(context.Response.Header.Peek(fasthttp.HeaderLastModified)); err == nil {
		return lastModified
	} else {
		return time.Time{}
	}
}

func AddCacheControl(context *fasthttp.RequestCtx, value string) {
	context.Response.Header.Add(fasthttp.HeaderCacheControl, value)
}

func AddContentEncoding(context *fasthttp.RequestCtx, value string) {
	context.Response.Header.Add(fasthttp.HeaderContentEncoding, value)
}

func IfNoneMatch(context *fasthttp.RequestCtx, eTag string) bool {
	if eTag == "" {
		return false
	} else {
		ifNoneMatch := util.BytesToString(context.Request.Header.Peek(fasthttp.HeaderIfNoneMatch))
		return ifNoneMatch == eTag
	}
}

func NotFound(context *fasthttp.RequestCtx) bool {
	return context.Response.StatusCode() == fasthttp.StatusNotFound
}

// TODO: not good enough
func ParseAccept(context *Context) []string {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
	// TODO: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
	accept := strings.Split(util.BytesToString(context.context.Request.Header.Peek("Accept")), ",")
	//context.Log.Debugf("ACCEPT: %s", accept)
	return accept
}
