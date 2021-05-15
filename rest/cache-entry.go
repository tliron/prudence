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
	Headers    [][][]byte        // list of key, value tuples
	Body       map[string][]byte // encoding -> body
	Expiration time.Time
}

func NewCacheEntry(context *Context) *CacheEntry {
	body := make(map[string][]byte)
	if context.context.Request.Header.IsGet() {
		// Body exists only in GET
		body[GetContentEncoding(context.context)] = copyBytes(context.context.Response.Body())
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

func NewCacheBody(context *Context, encoding string, body []byte) *CacheEntry {
	return &CacheEntry{
		Body:       map[string][]byte{encoding: body},
		Headers:    nil,
		Expiration: time.Now().Add(time.Duration(context.CacheDuration * 1000000000.0)), // seconds to nanoseconds
	}
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
		var body []byte

		if context.context.Request.Header.HasAcceptEncoding("gzip") {
			body = self.GetBody("gzip")
		} else {
			body = self.GetBody("")
		}

		context.context.Response.SetBody(body)
	}
}

func (self *CacheEntry) Write(context *Context) (int, error) {
	body := self.GetBody("")
	return context.Write(body)
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

func (self *CacheEntry) GetBody(encoding string) []byte {
	var body []byte

	var ok bool
	switch encoding {
	case "gzip":
		if body, ok = self.Body["gzip"]; !ok {
			if plain, ok := self.Body[""]; ok {
				log.Debug("creating gzip body")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGzip(buffer, plain)
				body = buffer.Bytes()
				self.Body["gzip"] = body
			}
		}

	case "":
		if body, ok = self.Body[""]; !ok {
			if gzip, ok := self.Body["gzip"]; ok {
				log.Debug("creating plain body")
				buffer := bytes.NewBuffer(nil)
				fasthttp.WriteGunzip(buffer, gzip)
				body = buffer.Bytes()
				self.Body[""] = body
			}
		}
	}

	return body
}

func copyBytes(bytes []byte) []byte {
	bytes_ := make([]byte, len(bytes))
	copy(bytes_, bytes)
	return bytes_
}
