/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package http

import (
	"aelog"
	"cnst"
	"config"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"io"
	"log"
	"module"
	"net/http"
	"storage"
	"strconv"
	"strings"
	"temp"
	"time"
	"uploader"
)

const (
	ERROR_PAGE = "<html><head><title>%s</title></head><body><center><h1>%d %s</h1></center><hr><center>" + cnst.SIGN + "</center></body></html>\n"
)

type Server struct {
	stor *storage.Storage
	conf *config.Config
	aL   *aelog.AntLog
	up   *uploader.Uploader
}

// Create new server and run it
func RunServer(s *storage.Storage, accessLog *aelog.AntLog) (server *Server) {
	server = &Server{
		stor: s,
		conf: s.Conf,
		aL:   accessLog,
		up:   uploader.NewUploader(s.Conf, s),
	}
	module.RegisterModules()
	server.Run()
	return
}

// Run all servers
func (s *Server) Run() {
	run := func(handler fasthttp.RequestHandler, addr string) {
		log.Fatal(fasthttp.ListenAndServe(":80", handler))
	}
	if s.conf.HttpReadAddr != s.conf.HttpWriteAddr {
		go run(s.ReadOnly, s.conf.HttpReadAddr)
	}
	go run(s.ReadWrite, s.conf.HttpWriteAddr)
	return
}

// Http handler for read-only server
func (s *Server) ReadOnly(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Server", cnst.SIGN)
	defer func() {
		if rec := recover(); rec != nil {
			ctx.Response.Reset()
			s.Err(500, ctx)
			aelog.Warnf("Error on http request: %v", rec)
		}
	}()
	filename := Filename(string(ctx.Path()))
	if len(filename) == 0 {
		s.Err(404, ctx)
		return
	}

	switch {
	case ctx.IsGet():
		s.Get(filename, ctx, true)
		return
	case ctx.IsHead():
		s.Get(filename, ctx, false)
		return
	case string(ctx.Method()) == "OPTIONS":
		ctx.Response.Header.Set("Allow", "GET,HEAD")
		ctx.SetStatusCode(http.StatusOK)
		return
	default:
		s.Err(501, ctx)
	}
}

// Http handler for read-write server
func (s *Server) ReadWrite(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Server", cnst.SIGN)
	defer func() {
		if rec := recover(); rec != nil {
			ctx.Response.Reset()
			s.Err(500, ctx)
			aelog.Warnf("Error on http request: %v", rec)
		}
	}()
	filename := Filename(string(ctx.Path()))

	switch filename {
	case "":
		// check uploader
		/*isU, err, errCode := s.up.TryRequest(r, w)
		if isU {
			if err == nil && errCode > 0 {
				s.Err(errCode, r, w)
			}
			return
		}*/
		s.Err(404, ctx)
		return
	case s.conf.StatusHtml:
		s.Err(500, ctx)
		return
	case s.conf.StatusJson:
		s.StatsJson(ctx)
		return
	}

	m := string(ctx.Request.Header.Peek("X-Http-Method-Override"))
	if m == "" {
		m = string(ctx.Method())
	}
	sm := m

	du := s.downloadUrl(ctx)
	if du != "" {
		sm = "DOWNLOAD"
	}

	switch sm {
	case "GET":
		s.Get(filename, ctx, true)
		return
	case "HEAD":
		s.Get(filename, ctx, false)
		return
	case "POST":
		s.Save(filename, ctx)
		return
	case "PUT":
		s.Delete(filename, nil)
		s.Save(filename, ctx)
		return
	case "DELETE":
		s.Delete(filename, ctx)
		return
	case "DOWNLOAD":
		if m == "PUT" {
			s.Delete(filename, nil)
		}
		s.Download(filename, ctx)
		return
	case "COMMAND":
		s.Command(filename, ctx)
		return
	case "RENAME":
		s.Rename(filename, ctx)
		return
	case "OPTIONS":
		ctx.Response.Header.Set("Allow", "GET,HEAD,POST,PUT,DELETE")
		ctx.SetStatusCode(http.StatusOK)
		return
	default:
		s.Err(501, ctx)
	}
}

func (s *Server) Get(name string, ctx *fasthttp.RequestCtx, writeBody bool) {
	f, ok := s.stor.Get(name)
	if !ok {
		s.Err(404, ctx)
		s.stor.Stats.Counters.NotFound.Add()
		return
	}
	f.Open()
	defer f.Close()

	// Check cache
	cont, status := s.checkCache(ctx, f)
	if !cont {
		ctx.SetStatusCode(status)
		s.stor.Stats.Counters.NotModified.Add()
		s.accessLog(status, ctx)
		return
	}
	ctx.Response.Header.Set("Accept-Ranges", "bytes")
	ctx.Response.Header.Set("Content-Type", f.ContentType())
	ctx.Response.Header.Set("Content-Length", strconv.Itoa(int(f.FSize)))
	ctx.Response.Header.Set("Last-Modified", f.Time.UTC().Format(http.TimeFormat))
	if s.conf.ETagSupport {
		ctx.Response.Header.Set("E-Tag", f.ETag())
	}
	if s.conf.Md5Header {
		ctx.Response.Header.Set("X-Ae-Md5", f.Md5S())
	}

	// Add headers from config
	for k, v := range s.conf.Headers {
		ctx.Response.Header.Add(k, v)
	}

	s.stor.Stats.Counters.Get.Add()

	// check range request
	goServe := false
	partial := false
	ranges := string(ctx.Request.Header.Peek("Range"))
	if ranges != "" {
		if strings.TrimSpace(strings.ToLower(ranges)) == "bytes=0-" {
			partial = true
		} else {
			goServe = true
		}
	}

	// fix http go lib bug for a "0-" requests
	if partial {
		status = http.StatusPartialContent
		ctx.Response.Header.Add("Content-Range", fmt.Sprintf("bytes 0-%d/%d", f.FSize-1, f.FSize))
	}

	if !writeBody {
		if status == http.StatusOK || status == http.StatusPartialContent {
			status = http.StatusNoContent
		}
		ctx.SetStatusCode(status)
		s.accessLog(status, ctx)
		return
	}

	reader := f.GetReader()

	if goServe {
		fasthttpadaptor.NewFastHTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeContent(w, r, name, f.Time, reader)
		})(ctx)

	} else {
		ctx.SetStatusCode(status)
		reader.WriteTo(ctx)
	}

	s.accessLog(status, ctx)
}

func (s *Server) Save(name string, ctx *fasthttp.RequestCtx) {
	_, ok := s.stor.Get(name)
	if ok {
		// File exists
		s.Err(409, ctx)
		return
	}

	//size := ctx.Request.Header.ContentLength()
	//s.save(name, size, ctx, ctx)
}

func (s *Server) Download(name string, ctx *fasthttp.RequestCtx) {
	_, ok := s.stor.Get(name)
	if ok {
		// File exists
		s.Err(409, ctx)
		return
	}
	url := s.downloadUrl(ctx)
	tf := temp.NewFile(s.conf.TmpDir)
	aelog.Debugf("Start download from %s\n", url)
	err := tf.LoadFromUrl(url)
	defer tf.Close()
	if err != nil {
		aelog.Infof("Can't download : %s, err: %v\n", url, err)
		s.Err(500, ctx)
		return
	}
	// again check for exists
	_, ok = s.stor.Get(name)
	if ok {
		s.Err(409, ctx)
		return
	}
	s.save(name, tf.Size, tf.File, ctx)
}

func (s *Server) Command(name string, ctx *fasthttp.RequestCtx) {
	/*
		command := strings.ToLower(r.Header.Get("X-Ae-Command"))
		if cont, _ := module.OnCommand(command, name, w, r, s.stor); cont {
			s.Get(name, w, r, false)
		}
	*/
	ctx.SetStatusCode(http.StatusMethodNotAllowed)
}

func (s *Server) Rename(name string, ctx *fasthttp.RequestCtx) {
	_, ok := s.stor.Index.Get(name)
	if !ok {
		s.Err(http.StatusNotFound, ctx)
		return
	}
	newName := strings.Trim(string(ctx.Request.Header.Peek("X-Ae-Name")), "/")
	if newName == "" {
		s.Err(http.StatusBadRequest, ctx)
		return
	}

	forceString := strings.TrimSpace(string(ctx.Request.Header.Peek("X-Ae-Force")))
	force := false
	switch strings.ToLower(forceString) {
	case "1", "true":
		force = true
	}
	if _, ok = s.stor.Get(newName); ok {
		if !force {
			s.Err(http.StatusConflict, ctx)
			return
		} else {
			s.stor.Delete(newName)
		}
	}

	_, err := s.stor.Index.Rename(name, newName)
	switch err {
	case storage.ErrFileNotFound:
		s.Err(http.StatusNotFound, ctx)
		return
	case storage.ErrConflict:
		s.Err(http.StatusConflict, ctx)
		return
	default:
		if err != nil {
			aelog.Warnf("Can't rename file: %v", err)
			s.Err(http.StatusInternalServerError, ctx)
			return
		}
	}
	s.Get(newName, ctx, false)
	return
}

func (s *Server) save(name string, size int64, reader io.Reader, ctx *fasthttp.RequestCtx) {
	if size <= 0 {
		s.Err(411, ctx)
		return
	}
	if size > s.conf.ContainerSize {
		s.Err(413, ctx)
		return
	}
	f, err := s.stor.Add(name, reader, size)
	if err != nil {
		s.Err(500, ctx)
		return
	}
	ctx.Response.Header.Set("X-Ae-Md5", f.Md5S())
	ctx.Response.Header.Set("Etag", f.ETag())
	ctx.Response.Header.Set("Location", name)
	/*
		if err = module.OnSave(f, w, r, s.stor); err != nil {
			s.Err(500, r, w)
			return
		}*/
	ctx.SetStatusCode(http.StatusCreated)
	s.accessLog(http.StatusCreated, ctx)
}

func (s *Server) Delete(name string, ctx *fasthttp.RequestCtx) {
	var ok bool
	var mode string
	if ctx != nil {
		mode = strings.ToLower(string(ctx.Request.Header.Peek("X-Ae-Delete")))
	}
	switch mode {
	case "childs":
		ok = s.stor.DeleteChilds(name)
	case "all":
		cok := s.stor.DeleteChilds(name)
		ok = s.stor.Delete(name) || cok
	default:
		ok = s.stor.Delete(name)
	}

	if ok {
		if ctx != nil {
			ctx.SetStatusCode(http.StatusNoContent)
			s.accessLog(http.StatusNoContent, ctx)
			s.stor.Stats.Counters.Delete.Add()
		}
		return
	} else {
		if ctx != nil {
			s.Err(404, ctx)
		}
	}
}

func (s *Server) StatsJson(ctx *fasthttp.RequestCtx) {
	b := s.stor.GetStats().AsJson()
	ctx.Response.Header.SetContentType("application/json;charset=utf-8")
	ctx.Write(b)
	s.accessLog(http.StatusOK, ctx)
}

func (s *Server) Err(code int, ctx *fasthttp.RequestCtx) {
	st := http.StatusText(code)
	body := []byte(fmt.Sprintf(ERROR_PAGE, st, code, st))
	ctx.SetBody(body)
	ctx.Response.Header.SetContentType("text/html;charset=utf-8")
	ctx.Response.Header.SetContentLength(len(body))
	ctx.SetStatusCode(code)
	s.accessLog(code, ctx)
}

/**
 * Return slash-trimmed filename
 */
func Filename(path string) string {
	return strings.Trim(path, "/")
}

/**
 * Detect download url and return if found
 */
func (s *Server) downloadUrl(ctx *fasthttp.RequestCtx) (url string) {
	if s.conf.DownloaderEnable {
		p := s.conf.DownloaderParamName
		url = string(ctx.FormValue(p))
	}
	return
}

func (s *Server) checkCache(ctx *fasthttp.RequestCtx, f *storage.File) (cont bool, status int) {
	// Check etag
	if s.conf.ETagSupport {
		if ifNoneMatch := string(ctx.Request.Header.Peek("If-None-Match")); ifNoneMatch != "" {
			if ifNoneMatch == f.ETag() {
				status = http.StatusNotModified
				return
			}
		}
	}
	// Check if modified
	if tm, err := time.Parse(http.TimeFormat, string(ctx.Request.Header.Peek("If-Modified-Since"))); err == nil && f.Time.Before(tm.Add(1*time.Second)) {
		status = http.StatusNotModified
		return
	}
	cont = true
	status = http.StatusOK
	return
}

func (s *Server) accessLog(status int, ctx *fasthttp.RequestCtx) {
	if s.aL != nil {
		st := http.StatusText(status)
		s.aL.Printf(aelog.LOG_PRINT, "%s %s (%s): %d %s", string(ctx.Method()), string(ctx.Path()), ctx.RemoteIP().String(), status, st)
	}
}
