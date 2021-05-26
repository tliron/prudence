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
	if context.Context.Request.Header.IsGet() {
		// Body exists only in GET
		contentEncoding := GetContentEncoding(context.Context)
		if encodingType := GetEncodingType(contentEncoding); encodingType != platform.EncodingTypeUnsupported {
			body[encodingType] = copyBytes(context.Context.Response.Body())
		} else {
			log.Warningf("unsupported encoding: %s", contentEncoding)
		}
	}

	// This is an annoying way to get all headers, but unfortunately if we
	// get the entire header via Header() there is no API to set it correctly
	// in CacheEntry.Write
	var headers [][][]byte
	context.Context.Response.Header.VisitAll(func(key []byte, value []byte) {
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

func CacheEntryGetBestBody(cacheEntry *platform.CacheEntry, context *Context) ([]byte, bool) {
	if context.Context.Request.Header.HasAcceptEncoding("br") {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypeBrotli)
	} else if context.Context.Request.Header.HasAcceptEncoding("gzip") {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypeGZip)
	} else if context.Context.Request.Header.HasAcceptEncoding("deflate") {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypeDeflate)
	} else {
		return CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain)
	}
}

func CacheEntryGetBody(cacheEntry *platform.CacheEntry, encoding platform.EncodingType) ([]byte, bool) {
	if body, ok := cacheEntry.Body[encoding]; ok {
		return body, false
	} else {
		switch encoding {
		case platform.EncodingTypeBrotli:
			if plain, _ := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); plain != nil {
				log.Debug("creating brotli body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteBrotli(buffer, plain)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypeBrotli] = body
				return body, true
			}

		case platform.EncodingTypeGZip:
			if plain, _ := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); plain != nil {
				log.Debug("creating gzip body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGzip(buffer, plain)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypeGZip] = body
				return body, true
			}

		case platform.EncodingTypeDeflate:
			if plain, _ := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); plain != nil {
				log.Debug("creating deflate body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteDeflate(buffer, plain)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypeDeflate] = body
				return body, true
			}

		case platform.EncodingTypePlain:
			// Try decoding an existing body
			if deflate, ok := cacheEntry.Body[platform.EncodingTypeDeflate]; ok {
				log.Debug("creating plain body from default")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteInflate(buffer, deflate)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypePlain] = body
				return body, true
			} else if gzip, ok := cacheEntry.Body[platform.EncodingTypeGZip]; ok {
				log.Debug("creating plain body from gzip")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGunzip(buffer, gzip)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypePlain] = body
				return body, true
			} else if brotli, ok := cacheEntry.Body[platform.EncodingTypeBrotli]; ok {
				log.Debug("creating plain body from brotli")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteUnbrotli(buffer, brotli)
				body = buffer.Bytes()
				cacheEntry.Body[platform.EncodingTypePlain] = body
				return body, true
			}
		}

		return nil, false
	}
}

func CacheEntryDescribe(cacheEntry *platform.CacheEntry, context *Context) bool {
	context.Context.Response.Reset()

	// Annoyingly these were re-enabled by Reset above
	context.Context.Response.Header.DisableNormalizing()
	context.Context.Response.Header.SetNoDefaultContentType(true)

	if context.Debug {
		context.Context.Response.Header.Set(CACHED_HEADER, context.CacheKey)
	}

	for _, header := range cacheEntry.Headers {
		context.Context.Response.Header.AddBytesKV(header[0], header[1])
	}

	return !context.isNotModified()
}

func CacheEntryPresent(cacheEntry *platform.CacheEntry, context *Context) bool {
	if !CacheEntryDescribe(cacheEntry, context) {
		return false
	}

	// Match client-side caching with server-side caching
	maxAge := int(cacheEntry.TimeToLive())
	AddCacheControl(context.Context, fmt.Sprintf("max-age=%d", maxAge))

	if context.Context.IsGet() {
		body, changed := CacheEntryGetBestBody(cacheEntry, context)
		context.Context.Response.SetBody(body)
		return changed
	}

	return false
}

func CacheEntryWrite(cacheEntry *platform.CacheEntry, context *Context) (bool, int, error) {
	if body, changed := CacheEntryGetBody(cacheEntry, platform.EncodingTypePlain); body != nil {
		n, err := context.Write(body)
		return changed, n, err
	} else {
		return false, 0, nil
	}
}

// Util

func copyBytes(bytes []byte) []byte {
	bytes_ := make([]byte, len(bytes))
	copy(bytes_, bytes)
	return bytes_
}
