package rest

import (
	"bytes"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

//
// CacheEntry
//

type CacheEntry struct {
	Headers    [][][]byte              // list of key, value tuples
	Body       map[EncodingType][]byte // encoding type -> body
	Expiration time.Time
}

func NewCacheEntry(context *Context) *CacheEntry {
	body := make(map[EncodingType][]byte)
	if context.context.Request.Header.IsGet() {
		// Body exists only in GET
		contentEncoding := GetContentEncoding(context.context)
		if encodingType := GetEncodingType(contentEncoding); encodingType != EncodingTypeUnsupported {
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

	return &CacheEntry{
		Body:       body,
		Headers:    headers,
		Expiration: time.Now().Add(time.Duration(context.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

func NewCacheEntryBody(context *Context, encoding EncodingType, body []byte) *CacheEntry {
	return &CacheEntry{
		Body:       map[EncodingType][]byte{encoding: body},
		Headers:    nil,
		Expiration: time.Now().Add(time.Duration(context.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
}

// fmt.Stringer interface
func (self *CacheEntry) String() string {
	keys := make([]string, 0, len(self.Body))
	for key := range self.Body {
		keys = append(keys, key.String())
	}
	return fmt.Sprintf("%s", keys)
}

func (self *CacheEntry) Expired() bool {
	return time.Now().After(self.Expiration)
}

// In seconds
func (self *CacheEntry) TimeToLive() float64 {
	duration := self.Expiration.Sub(time.Now()).Seconds()
	if duration < 0.0 {
		duration = 0.0
	}
	return duration
}

func (self *CacheEntry) GetBestBody(context *Context) []byte {
	if context.context.Request.Header.HasAcceptEncoding("br") {
		return self.GetBody(EncodingTypeBrotli)
	} else if context.context.Request.Header.HasAcceptEncoding("gzip") {
		return self.GetBody(EncodingTypeGZip)
	} else if context.context.Request.Header.HasAcceptEncoding("deflate") {
		return self.GetBody(EncodingTypeDeflate)
	} else {
		return self.GetBody(EncodingTypePlain)
	}
}

func (self *CacheEntry) GetBody(type_ EncodingType) []byte {
	var body []byte

	// TODO: we need to update the backend if we change the entry!

	var ok bool
	if body, ok = self.Body[type_]; !ok {
		switch type_ {
		case EncodingTypeBrotli:
			if plain := self.GetBody(EncodingTypePlain); plain != nil {
				log.Debug("creating brotli body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteBrotli(buffer, plain)
				body = buffer.Bytes()
				self.Body[EncodingTypeBrotli] = body
			}

		case EncodingTypeGZip:
			if plain := self.GetBody(EncodingTypePlain); plain != nil {
				log.Debug("creating gzip body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGzip(buffer, plain)
				body = buffer.Bytes()
				self.Body[EncodingTypeGZip] = body
			}

		case EncodingTypeDeflate:
			if plain := self.GetBody(EncodingTypePlain); plain != nil {
				log.Debug("creating deflate body from plain")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteDeflate(buffer, plain)
				body = buffer.Bytes()
				self.Body[EncodingTypeDeflate] = body
			}

		case EncodingTypePlain:
			// Try decoding an existing body
			if deflate, ok := self.Body[EncodingTypeDeflate]; ok {
				log.Debug("creating plain body from default")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteInflate(buffer, deflate)
				body = buffer.Bytes()
				self.Body[EncodingTypePlain] = body
			} else if gzip, ok := self.Body[EncodingTypeGZip]; ok {
				log.Debug("creating plain body from gzip")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGunzip(buffer, gzip)
				body = buffer.Bytes()
				self.Body[EncodingTypePlain] = body
			} else if brotli, ok := self.Body[EncodingTypeBrotli]; ok {
				log.Debug("creating plain body from brotli")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteUnbrotli(buffer, brotli)
				body = buffer.Bytes()
				self.Body[EncodingTypePlain] = body
			}
		}
	}

	return body
}

func (self *CacheEntry) ToContext(context *Context) {
	context.context.Response.Reset()

	// Annoyingly these were re-enabled by Reset above
	context.context.Response.Header.DisableNormalizing()
	context.context.Response.Header.SetNoDefaultContentType(true)

	// Headers
	for _, header := range self.Headers {
		context.context.Response.Header.AddBytesKV(header[0], header[1])
	}

	eTag := GetETag(context.context)

	// New max-age
	maxAge := int(self.TimeToLive())
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
		body := self.GetBestBody(context)
		context.context.Response.SetBody(body)
	}
}

func (self *CacheEntry) Write(context *Context) (int, error) {
	if body := self.GetBody(EncodingTypePlain); body != nil {
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
