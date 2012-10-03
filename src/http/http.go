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
	"cnst"
	"storage"
	"config"
	"net/http"
	"log"
	"fmt"
	"time"
	"strconv"
	"io"
	"aelog"
	"strings"
	"temp"
	"uploader"
)

const (
	ERROR_PAGE = "<html><head><title>%s</title></head><body><center><h1>%d %s</h1></center><hr><center>" + cnst.SIGN + "</center></body></html>\n"
)


type Server struct {
	stor *storage.Storage
	conf *config.Config
	aL *aelog.AntLog
	up *uploader.Uploader
}

// Create new server and run it
func RunServer(s *storage.Storage, accessLog *aelog.AntLog) (server *Server) {
	server = &Server{
		stor : s,
		conf : s.Conf,
		aL   : accessLog,
		up   : uploader.NewUploader(s.Conf, s),
	}
	server.Run()
	return
}

// Run all servers
func (s *Server) Run() {
	run := func(handler http.Handler, addr string) { 
		serv := &http.Server{
			Addr:         addr,
			Handler:      handler,
		}
		log.Fatal(serv.ListenAndServe())
	}
	if s.conf.HttpReadAddr != s.conf.HttpWriteAddr {
		go run(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.ReadOnly(w, r)
		}), s.conf.HttpReadAddr)
	}
	go run(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.ReadWrite(w, r)
	}), s.conf.HttpWriteAddr)
	return
}

// Http handler for read-only server
func (s *Server) ReadOnly(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", cnst.SIGN)
	defer func() {
		if rec := recover(); rec != nil {
			s.Err(500, r, w)
        	aelog.Warnf("Error on http request: %v", rec)
        }
        r.Body.Close()
	}()
	filename := Filename(r)
	if len(filename) == 0 {
		s.Err(404, r, w)
		return 
	}
	
	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Allow", "GET,HEAD")
		w.WriteHeader(http.StatusOK)
		return
	case "GET":
		s.Get(filename, w, r, true)
		return
	case "HEAD":
		s.Get(filename, w, r, false)
		return
	default:
		s.Err(501, r, w)
	}
}

// Http handler for read-write server
func (s *Server) ReadWrite(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", cnst.SIGN)
	defer func() {
		if rec := recover(); rec != nil {			
        	s.Err(500, r, w)
        	aelog.Warnf("Error on http request: %v", rec)
        }
        r.Body.Close()
	}()
	filename := Filename(r)
	switch filename {
		case "":
			// check uploader
			isU, err, errCode := s.up.TryRequest(r, w)
			if isU {
				if err == nil && errCode > 0 {
					s.Err(errCode, r, w)
				}
				return
			}
			s.Err(404, r, w)
			return;
		case s.conf.StatusHtml:
			s.Err(500, r, w)
			return
		case s.conf.StatusJson:
			s.StatsJson(w, r)
			return
	}
	
	m := r.Header.Get("X-Http-Method-Override")
	if m == "" {
		m = r.Method
	}
	sm := m
	
	du := s.downloadUrl(r)
	if du != "" { 
		sm = "DOWNLOAD"
	}
	
	switch sm {
	case "OPTIONS":
		w.Header().Set("Allow", "GET,HEAD,POST,PUT,DELETE")
		w.WriteHeader(http.StatusOK)
		return
	case "GET":
		s.Get(filename, w, r, true)
		return
	case "HEAD":
		s.Get(filename, w, r, false)
		return
	case "POST":
		s.Save(filename, w, r)
		return
	case "PUT":
		s.Delete(filename, nil, nil)
		s.Save(filename, w, r)
		return
	case "DELETE":
		s.Delete(filename, w, r)
		return
	case "DOWNLOAD":
		if m == "PUT" {
			s.Delete(filename, nil, nil)
		}
		s.Download(filename, w, r)
		return
	default:
		s.Err(501, r, w)
	}
}

func (s *Server) Get(name string, w http.ResponseWriter, r *http.Request, writeBody bool) {
	f, ok := s.stor.Get(name)
	if ! ok {
		s.Err(404, r, w)
		s.stor.Stats.Counters.NotFound.Add()
		return
	}
	f.Open()
	defer f.Close()
	
	// Check cache
	cont, status := s.checkCache(r, f)
	if ! cont {
		w.WriteHeader(status)
		s.stor.Stats.Counters.NotModified.Add()
		s.accessLog(status, r)
		return
	}
	w.Header().Set("Content-Type", f.ContentType())
	w.Header().Set("Content-Length", strconv.Itoa(int(f.Size)))
	w.Header().Set("Last-Modified", f.Time.UTC().Format(http.TimeFormat))
	if s.conf.ETagSupport {
		w.Header().Set("E-Tag", f.ETag())
	}
	
	
	// Add headers from config
	for k, v := range s.conf.Headers {
		w.Header().Add(k, v)
	}
	
	s.stor.Stats.Counters.Get.Add()
	
	if ! writeBody {
		w.WriteHeader(status)
		s.accessLog(status, r)
		return
	}
	
	reader := f.GetReader()
	
	if f.Size > s.conf.ContentRange {
		http.ServeContent(w, r, name, f.Time, reader)
	} else {
		io.Copy(w, reader)
	}
	
	s.accessLog(200, r)
}

func (s *Server) Save(name string, w http.ResponseWriter, r *http.Request) {
	_, ok := s.stor.Get(name)
	if ok {
		// File exists
		s.Err(409, r, w)
		return
	}
	
	reader := r.Body
	size := r.ContentLength
	s.save(name, size, reader, r, w)
}


func (s *Server) Download(name string, w http.ResponseWriter, r *http.Request) {
	_, ok := s.stor.Get(name)
	if ok {
		// File exists
		s.Err(409, r, w)
		return
	}
	url := s.downloadUrl(r)
	tf := temp.NewFile(s.conf.TmpDir)
	aelog.Debugf("Start download from %s\n", url)
	err := tf.LoadFromUrl(url)
	if err != nil {
		aelog.Infof("Can't download : %s, err: %v\n", url, err)
		s.Err(500, r, w)
		return
	}
	defer tf.Close()
	s.save(name, tf.Size, tf.File, r, w)
}


func (s *Server) save(name string, size int64, reader io.Reader, r *http.Request, w http.ResponseWriter) {
	if size <= 0 {
		s.Err(411, r, w)
		return
	}	
	if size > s.conf.ContainerSize {
		s.Err(413, r, w)
		return
	}
	s.stor.Stats.Counters.Add.Add()
	f := s.stor.Add(name, reader, size)
	w.Header().Set("X-Ae-Md5", fmt.Sprintf("%x", f.Md5));	
	w.Header().Set("Etag", f.ETag());
	w.Header().Set("Location", name);
	w.WriteHeader(http.StatusCreated)
	s.accessLog(http.StatusCreated, r)
}

func (s *Server) Delete(name string, w http.ResponseWriter, r *http.Request) {
	ok := s.stor.Delete(name)
	if ok {
		if w != nil {
			w.WriteHeader(http.StatusNoContent)
			s.accessLog(http.StatusCreated, r)
			s.stor.Stats.Counters.Delete.Add()
		}		
		return
	} else {
		if w != nil {
			s.Err(404, r, w)
		}
	}
}

func (s *Server) StatsJson(w http.ResponseWriter, r *http.Request) {
	b := s.stor.GetStats().AsJson()
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(b)
	s.accessLog(http.StatusOK, r)
}


func (s *Server) Err(code int, r *http.Request, w http.ResponseWriter) {
	st := http.StatusText(code)
	body := []byte(fmt.Sprintf(ERROR_PAGE, st, code, st))
	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	w.Header().Add("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(code)
	w.Write(body)
	s.accessLog(code, r)
}


/**
 * Return slash-trimmed filename
 */
func Filename(r *http.Request) (fn string) {
	fn = r.URL.Path
	fn = strings.Trim(fn, "/")
	return
}


/**
 * Detect download url and return if found
 */
func (s *Server) downloadUrl(r *http.Request) (url string) {
	if s.conf.DownloaderEnable {
		p := s.conf.DownloaderParamName
		url = r.FormValue(p)
		if url == "" {
			url = r.URL.Query().Get(p)
		}
	}
	return
}

func (s *Server) checkCache(r *http.Request, f *storage.File) (cont bool, status int) {
	// Check etag
	if s.conf.ETagSupport {
		if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
			if ifNoneMatch == f.ETag() {
				status = http.StatusNotModified
				return
			}
		}
	}	
	// Check if modified
	if tm, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && f.Time.Before(tm.Add(1*time.Second)) {
		status = http.StatusNotModified
		return
   	}
   	cont = true
   	status = http.StatusOK
	return
}

func (s *Server) accessLog(status int, r *http.Request) {
	if s.aL != nil {
		st := http.StatusText(status)
		s.aL.Printf(aelog.LOG_PRINT, "%s %s (%s): %s", r.Method, r.URL.Path, r.RemoteAddr, st)
	}
}

