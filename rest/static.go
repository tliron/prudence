package rest

import (
	"bytes"
	contextpkg "context"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tliron/commonjs-goja"
	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

//
// Static
//

type Static struct {
	Root               string
	Indexes            []string
	PresentDirectories bool
}

func NewStatic(root string, indexes ...string) *Static {
	// TODO: support indexes
	return &Static{
		Root:    root,
		Indexes: indexes,
	}
}

// ([platform.CreateFunc] signature)
func CreateStatic(jsContext *commonjs.Context, config ard.StringMap) (any, error) {
	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	var self Static

	self.Root, _ = config_.Get("root").String()
	if rootUrl, err := jsContext.Resolve(contextpkg.TODO(), self.Root, true); err == nil {
		if rootFileUrl, ok := rootUrl.(*exturl.FileURL); ok {
			self.Root = rootFileUrl.Path
		} else {
			return nil, fmt.Errorf("Static \"root\" is not a file: %v", rootUrl)
		}
	} else {
		return nil, err
	}

	self.Indexes = platform.AsStringList(config_.Get("indexes"))

	if presentDirectories, ok := config_.Get("presentDirectories").Boolean(); ok {
		self.PresentDirectories = presentDirectories
	}

	return &self, nil
}

// ([Handler] interface, [HandleFunc] signature)
func (self *Static) Handle(restContext *Context) (handled bool, rerr error) {
	path := filepath.Join(self.Root, restContext.Request.Path)

	/*
		// NOTE: ServeFile annoyingly redirect "index.html" and also sets the response status.
		// The latter unwanted feature can be fixed with the ResponseWriter, but the former
		// cannot be fixed.

		responseWriter := restContext.Response.NewResponseWriterWrapper()
		http.ServeFile(responseWriter, restContext.Request.Direct, path)
		if restContext.Response.Status != http.StatusNotFound {
			restContext.Response.Bypass = true
			return true, nil
		} else {
			return false, nil
		}
	*/

	handleFileError := func(err error) (bool, error) {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		if errors.Is(err, fs.ErrPermission) {
			restContext.Response.Status = http.StatusForbidden
			return true, nil
		}

		return false, err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return handleFileError(err)
	}

	if fileInfo.IsDir() {
		// Try indexes
		var found bool
		for _, index := range self.Indexes {
			path_ := filepath.Join(path, index)
			if fileInfo_, err := os.Stat(path_); err == nil {
				path = path_
				fileInfo = fileInfo_
				found = true
				break
			}
		}

		if !found && !self.PresentDirectories {
			return false, nil
		}
	}

	switch restContext.Request.Method {
	case "GET":
		file, err := os.Open(path)
		if err != nil {
			return handleFileError(err)
		}
		defer func() {
			rerr = file.Close()
		}()

		if fileInfo.IsDir() {
			if self.PresentDirectories {
				return true, PresentDirectory(fileInfo, file, restContext)
			} else {
				return false, nil
			}
		}

		http.ServeContent(restContext.Response.Direct, restContext.Request.Direct, path, fileInfo.ModTime(), file)

	case "HEAD":
		SetLastModifiedHeader(restContext.Response.Header, fileInfo.ModTime())
		if contentType := mime.TypeByExtension(filepath.Ext(path)); contentType != "" {
			restContext.Response.Header.Set(HeaderContentType, contentType)
		}

	default:
		restContext.Response.Status = http.StatusMethodNotAllowed
	}

	return true, nil
}

func PresentDirectory(fileInfo fs.FileInfo, file *os.File, restContext *Context) error {
	if err := restContext.RedirectTrailingSlash(0); err != nil {
		return err
	}

	serverTimestamp := fileInfo.ModTime()
	if clientTimestamp, ok := GetTimeHeader(HeaderIfModifiedSince, restContext.Request.Header); ok {
		if !serverTimestamp.Truncate(time.Second).After(clientTimestamp) {
			restContext.Response.Status = http.StatusNotModified
			restContext.Log.Debug("not modified: Last-Modified")
			return nil
		}
	}

	// TODO: check if modified

	if dirEntries, err := file.ReadDir(-1); err == nil {
		title := "/" + html.EscapeString(restContext.Request.Path)

		buffer := bytes.NewBuffer(nil)

		buffer.WriteString(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<link rel="icon" href="data:," />
<title>`)
		buffer.WriteString(title)
		buffer.WriteString(`</title>
</head>
<body>
<pre>
<h1>`)
		buffer.WriteString(title)
		buffer.WriteString("</h1>\n")

		for _, dirEntry := range dirEntries {
			name := html.EscapeString(dirEntry.Name())
			if dirEntry.IsDir() {
				name += "/"
			}

			buffer.WriteString(`<a href="`)
			buffer.WriteString(name)
			buffer.WriteString(`">`)
			buffer.WriteString(name)
			buffer.WriteString("</a>\n")
		}

		buffer.WriteString(`</pre>
</body>
</html>
`)

		SetLastModifiedHeader(restContext.Response.Header, serverTimestamp)
		SetContentTypeHeader(restContext.Response.Header, "text/html", "utf-8")

		return restContext.Write(buffer.Bytes())
	} else {
		return err
	}
}
