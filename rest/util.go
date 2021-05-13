package rest

import (
	"time"

	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

func GetResponseETag(context *fasthttp.RequestCtx) string {
	return util.BytesToString(context.Response.Header.Peek(fasthttp.HeaderETag))
}

func GetResponseLastModified(context *fasthttp.RequestCtx) time.Time {
	if lastModified, err := fasthttp.ParseHTTPDate(context.Response.Header.Peek(fasthttp.HeaderLastModified)); err == nil {
		return lastModified
	} else {
		return time.Time{}
	}
}

func AddCacheControl(context *fasthttp.RequestCtx, value string) {
	context.Response.Header.Add(fasthttp.HeaderCacheControl, value)
}

func IfNoneMatch(context *fasthttp.RequestCtx, eTag string) bool {
	if eTag == "" {
		return false
	} else {
		ifNoneMatch := util.BytesToString(context.Request.Header.Peek(fasthttp.HeaderIfNoneMatch))
		return ifNoneMatch == eTag
	}
}
