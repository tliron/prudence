package rest

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tliron/prudence/platform"
	"github.com/valyala/fasthttp"
)

func NewCacheEntryFromContext(context *Context) *platform.CacheEntry {
	body := make(map[platform.EncodingType][]byte)
	if context.context.Request.Header.IsGet() {
		// Body exists only in GET
		contentEncoding := GetContentEncoding(context.context)
		if encodingType := GetEncodingType(contentEncoding); encodingType != platform.EncodingTypeUnsupported {
			body[encodingType] = copyBytes(context.context.Response.Body())
		} else {
			log.Warningf("unsupported encoding: %s", contentEncoding)
		}
	}

	// This is an annoying way to get all headers, but unfortunately if we
	// get the entire header via Header() there is no API to set it correctly
	// in CacheEntry.Write
	var headers [][][]byte
	context.context.Response.Header.VisitAll(func(key []byte, value []byte) {
		switch string(key) {
		case fasthttp.HeaderServer, fasthttp.HeaderCacheControl:
			return
		}

		//context.Log.Debugf("header: %s", key)
		headers = append(headers, [][]byte{copyBytes(key), copyBytes(value)})
	})

	return &platform.CacheEntry{
		Body:       body,
		Headers:    headers,
		Expiration: time.Now().Add(time.Duration(context.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func NewCacheEntryFromBody(context *Context, encoding platform.EncodingType, body []byte) *platform.CacheEntry {
	return &platform.CacheEntry{
		Body:       map[platform.EncodingType][]byte{encoding: body},
		Headers:    nil,
		Expiration: time.Now().Add(time.Duration(context.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func CacheEntryGetBestBody(cacheEntry *platform.CacheEntry, context *Context) []byte {
	if context.context.Request.Header.HasAcceptEncoding("br") {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypeBrotli)
	} else if context.context.Request.Header.HasAcceptEncoding("gzip") {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypeGZip)
	} else if context.context.Request.Header.HasAcceptEncoding("deflate") {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypeDeflate)
	} else {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain)
	}
}

func CacheEntryGetBody(cacheEntry *platform.CacheEntry, encoding platform.EncodingType) []byte {
	var body []byte

	// TODO: we need to update the backend if we change the entry!

	var ok bool
	if body, ok = cacheEntry.Body[encoding]; !ok {
		switch encoding {
		case platform.EncodingTypeBrotli:
			if plain := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); plain != nil {
				log.Debug("creating brotli body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteBrotli(buffer, plain)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypeBrotli] = body
			}

		case platform.EncodingTypeGZip:
			if plain := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); plain != nil {
				log.Debug("creating gzip body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGzip(buffer, plain)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypeGZip] = body
			}

		case platform.EncodingTypeDeflate:
			if plain := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); plain != nil {
				log.Debug("creating deflate body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteDeflate(buffer, plain)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypeDeflate] = body
			}

		case platform.EncodingTypePlain:
			// Try decoding an existing body
			if deflate, ok := cacheEntry.Body[platform.EncodingTypeDeflate]; ok {
				log.Debug("creating plain body from default")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteInflate(buffer, deflate)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypePlain] = body
			} else if gzip, ok := cacheEntry.Body[platform.EncodingTypeGZip]; ok {
				log.Debug("creating plain body from gzip")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGunzip(buffer, gzip)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypePlain] = body
			} else if brotli, ok := cacheEntry.Body[platform.EncodingTypeBrotli]; ok {
				log.Debug("creating plain body from brotli")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteUnbrotli(buffer, brotli)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypePlain] = body
			}
		}
	}

	return body
}

func CacheEntryToContext(cacheEntry *platform.CacheEntry, context *Context) {
	context.context.Response.Reset()

	// Annoyingly these were re-enabled by Reset above
	context.context.Response.Header.DisableNormalizing()
	context.context.Response.Header.SetNoDefaultContentType(true)

	// Headers
	for _, header := range cacheEntry.Headers {
		context.context.Response.Header.AddBytesKV(header[0], header[1])
	}

	eTag := GetETag(context.context)

	// New max-age
	maxAge := int(cacheEntry.TimeToLive())
	AddCacheControl(context.context, fmt.Sprintf("max-age=%d", maxAge))

	// TODO only for debug mode
	context.context.Response.Header.Set("X-Prudence-Cached", context.CacheKey)

	// Conditional

	if IfNoneMatch(context.context, eTag) {
		// The following headers should have been set:
		// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
		context.context.NotModified()
		return
	}

	if !context.context.IfModifiedSince(GetLastModified(context.context)) {
		// The following headers should have been set:
		// Cache-Control, Content-Location, Date, ETag, Expires, and Vary
		context.context.NotModified()
		return
	}

	// Body (not for HEAD)

	if !context.context.IsHead() {
		body := CacheEntryGetBestBody(cacheEntry, context)
		context.context.Response.SetBody(body)
	}
}

func CacheEntryWritePlain(cacheEntry *platform.CacheEntry, context *Context) (int, error) {
	if body := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); body != nil {
		return context.Write(body)
	} else {
		return 0, nil
	}
}

// Util

func copyBytes(bytes []byte) []byte {
	bytes_ := make([]byte, len(bytes))
	copy(bytes_, bytes)
	return bytes_
}
