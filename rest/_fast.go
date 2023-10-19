package rest

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/go-scriptlet/jst"
	"github.com/tliron/kutil/util"
	"github.com/valyala/fasthttp"
)

func (self *Server) StartFast() error {
	self.log = commonlog.NewKeyValueLogger(log,
		"_scope", "server",
		"name", self.Name,
		"address", self.Address,
		"port", self.Port,
		"protocol", self.Protocol,
		"secure", self.TLS,
	)

	self.started.Add(1)
	defer self.started.Done()

	self.log.Info("starting")

	if listener, err := self.newListener(); err == nil {
		defer listener.Close()

		server := &fasthttp.Server{
			Handler: self.HandleFastHTTP,
		}

		if err := server.Serve(listener); (err == nil) || (err == http.ErrServerClosed) {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

// ([fasthttp.RequestHandler] signature)
func (self *Server) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	if self.Handler == nil {
		return
	}

	restContext := NewFastContext(ctx, self.log)

	defer func() {
		if r := recover(); r != nil {
			if r == EndRequest {
				restContext.Log.Debug("end")
				if err := restContext.Response.flush(); err != nil {
					self.log.Error(err.Error())
				}
			} else {
				panic(r)
			}
		}
	}()

	restContext.Debug = self.Debug

	if self.Name != "" {
		restContext.Response.StaticHeader.Set(HeaderServer, self.Name)
	}

	if _, err := self.Handler(restContext); err != nil {
		restContext.Response.Reset()
		if restContext.Debug {
			restContext.Write(err.Error())
			restContext.Write("\n")
		}
		restContext.InternalServerError(err)
	}

	if err := restContext.Response.flush(); err != nil {
		self.log.Error(err.Error())
	}
}

func NewFastContext(ctx *fasthttp.RequestCtx, log commonlog.Logger) *Context {
	id := requestId.Add(1)
	request_ := NewFastRequest(ctx)
	response := NewFastResponse(ctx)
	return &Context{
		Context:  jst.NewContext(response.Buffer, nil),
		Id:       id,
		Request:  request_,
		Response: response,
		Log: commonlog.NewKeyValueLogger(log,
			"_scope", "request",
			"request", id,
			"host", request_.Host,
			"port", request_.Port,
			"path", request_.Path,
		),
	}
}

func NewFastRequest(ctx *fasthttp.RequestCtx) *Request {
	host := util.BytesToString(ctx.Host())
	var port int
	if colon := strings.IndexRune(host, ':'); colon != -1 {
		port, _ = strconv.Atoi(host[colon+1:])
		host = host[:colon]
	}

	self := Request{
		Host: host,
		Port: port,
		Path: strings.TrimPrefix(util.BytesToString(ctx.URI().Path()), "/"),
		//Header:  request.Header.Clone(),
		Method: util.BytesToString(ctx.Method()),
		//Query:   request.URL.Query(), // will be a new map
		//Cookies: request.Cookies(),   // will be a new array
		//Direct:  request,
	}

	return &self
}

func NewFastResponse(ctx *fasthttp.RequestCtx) *Response {
	return &Response{
		Header:       make(http.Header),
		StaticHeader: make(http.Header),
		CharSet:      "utf-8",
		//Direct:       ctx,
		Buffer: bytes.NewBuffer(nil),
	}
}
