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

package anteater

import (
	"net/http"
	"fmt"
	"strconv"
	"io"
	"time"
	"path/filepath"
	"mime"
	"os"
	"crypto/md5"
)

const (
	ERROR_PAGE = "<html><head><title>%s</title></head><body><center><h1>%s</h1></center><hr><center>" + SERVER_SIGN + "</center></body></html>\n"
)


var httpErrors = map[int]string{
	400: "Invalid request",
	404: "404 Not Found",
	405: "405 Method Not Allowed",
	409: "409 Conflict",
	411: "411 Length Required",
	413: "413 Request Entity Too Large",
	500: "500 Internal Server Error",
	501: "501 Not Implemented",
}

/**
 * Start server with config params
 */
func RunServer(handler http.Handler, addr string) {
	s := &http.Server{
		Addr:         addr,
		Handler:      handler,
	}
	Log.Infof("Start http on %s ...\n", addr)
	Log.Fatal(s.ListenAndServe())
}

/**
 * Http read-only handler
 */
func HttpRead(w http.ResponseWriter, r *http.Request) {
		
	filename := GetFilename(r)
	
	if len(filename) == 0 {
		HttpError(w, 404)
		return
	}
	
	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Allow", "GET,HEAD")
		w.WriteHeader(http.StatusOK)
		return
	case "GET":
		HttpGetFile(filename, w, r, true)
		return
	case "HEAD":
		HttpGetFile(filename, w, r, false)
		return
	}
	HttpError(w, 501)
}

/**
 * Http read-write handler
 */
func HttpReadWrite(w http.ResponseWriter, r *http.Request) {

	filename := GetFilename(r)
	
	switch filename {
		case "":
			HttpError(w, 404)
			return
		case Conf.StatusJson:
			HttpPrintStatusJson(w)
			return
		case Conf.StatusHtml:
			HttpPrintStatusHtml(w)
			return
	}

	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Allow", "GET,HEAD,POST,PUT,DELETE")
		w.WriteHeader(http.StatusOK)
		return
	case "GET":
		HttpGetFile(filename, w, r, true)
		return
	case "HEAD":
		HttpGetFile(filename, w, r, false)
		return
	case "POST":
		HttpSaveFile(filename, w, r)
		return
	case "PUT":
		HttpDeleteFile(filename, nil)
		HttpSaveFile(filename, w, r)
		return
	case "DELETE":
		HttpDeleteFile(filename, w)
		return
	default:
		Log.Infoln("Unhandled http method", r.Method)
		HttpError(w, 501)
	}
}

/**
 * Return filename without first slashes
 */
func GetFilename(r *http.Request) string {
	var i int
	for _, s := range r.URL.Path {	
		if string(s) != "/" {
			break
		}
		i++
	}
	return r.URL.Path[i:]
}


/**
 * Http error handler
 */
func HttpError(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	switch status {
		case 405, 501:
			if w.Header().Get("Allow") == "" {
				w.Header().Set("Allow", "GET")
			}
		case 404:
			HttpCn.CNotFound()
	}

	fmt.Fprintf(w, ERROR_PAGE, httpErrors[status], httpErrors[status])
}

/**
 * Move file and headers to http response
 */
func HttpGetFile(name string, w http.ResponseWriter, r *http.Request, writeBody bool) {
	i, ok := IndexGet(name)
	if !ok {
		HttpError(w, 404)
		return
	}
	
	HttpCn.CGet()

	h, isContinue := httpHeadersHandle(name, i, w, r)
	
	for k, v := range(h) {
		w.Header().Set(k, v)
	}
	
	if  ! isContinue {
		return
	}
	
	// if need content-range support
	if i.Size > Conf.ContentRange {
		reader := i.GetReader()
		http.ServeContent(w, r, name, time.Unix(i.T, 0), reader)
		Log.Debugf("GET %s (%s) Size %d; Go Serve", r.URL, r.RemoteAddr)
		return
	}
	
	reader := i.GetReader()
	
	// if content type do not detected before
	if h["Content-Type"] == "" {
		// read a chunk to decide between utf-8 text and binary
		var buf [1024]byte
		n, _ := io.ReadFull(reader, buf[:])
		b := buf[:n]
		ctype := http.DetectContentType(b)
		_, err := reader.Seek(0, os.SEEK_SET) // rewind to output whole file
		if err != nil {
			Log.Warnln("Can't seek")
			return
		}
		w.Header().Set("Content-Type", ctype)
	}
	
	w.Header().Set("Content-Length", strconv.FormatInt(i.Size, 10))
	
	if ! writeBody {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// else just copy content to output
	n, err := io.Copy(w, reader)	
	if err != nil {
		Log.Warnf("GET %s (%s); Size: %d; Error! %v", r.URL, r.RemoteAddr, i.Size, err)
		HttpError(w, 500)
		return
	}
	if n != i.Size {
		Log.Warnf("GET %s (%s); Size: %d; Error! %s", r.URL, r.RemoteAddr, n, "Size not match")
		HttpError(w, 500)
		return
	}
	
	Log.Debugf("GET %s (%s); Size: %d; ", r.URL, r.RemoteAddr, n)
}

/**
 * Save file from http request
 */
func HttpSaveFile(name string, w http.ResponseWriter, r *http.Request) {
	_, ok := IndexGet(name)
	if ok {
		// File exists
		HttpError(w, 409)
		return
	}

	file := r.Body
	size := r.ContentLength
	
	if size == 0 {
		HttpError(w,  411)
		return
	}
	
	if size > Conf.ContainerSize {
		HttpError(w, 413)
		return
	}
	
	Log.Debugln("Start upload file", name, size, "bytes")
	
	f, err := GetFile(name, size)
	fi := f.Info()
	
	isOk := false

	defer func() {
		if isOk {
			IndexSet(name, fi)
		} else {
			FileContainers[fi.ContainerId].Delete(fi)
		}
	}()
	
	var written int64
	h := md5.New()
	for {
		buf := make([]byte, 256*1024)
		nr, er := io.ReadAtLeast(file, buf, 256*1024)
		if nr > 0 {
			nw, ew := f.WriteAt(buf[0:nr], written)
			if nw > 0 {
				written += int64(nw)
				h.Write(buf[0:nw])
			}
			if ew != nil {
				err = ew
				break
			}
		}
		if er != nil {
			err = er
			break
		}

		if err != nil {
			Log.Warnln(err)
			HttpError(w, 500)
			return
		}
	}
	
	md5 := h.Sum(nil)
	
	w.Header().Set("X-Ae-Md5", fmt.Sprintf("%x", md5));	
	w.Header().Set("Etag", fi.ETag());
	w.Header().Set("Location", name);
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(http.StatusCreated)
	HttpCn.CAdd()
	fi.Md5 = md5
	Log.Debugf("File %s (%d:%d) uploaded.\n", name, fi.ContainerId, fi.Id)
	isOk = true
}

/**
 * Delete file
 */
func HttpDeleteFile(name string, w http.ResponseWriter) {
	if i, ok := IndexDelete(name); ok {
		FileContainers[i.ContainerId].Delete(i)
		HttpCn.CDelete()
		if w != nil {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}
	if w != nil {
		HttpError(w, 404)
	}
}

/**
 * Handle http headers. Check cache, content-type, etc.
 */
func httpHeadersHandle(name string, i *FileInfo, w http.ResponseWriter, r *http.Request) (h map[string]string, isContinue bool) {
	// Check ETag
	if Conf.ETagSupport {
		if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
			if ifNoneMatch == i.ETag() {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}
	t := time.Unix(i.T, 0)
	
	// Check if modified
	if tm, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && t.Before(tm.Add(1*time.Second)) {
		w.WriteHeader(http.StatusNotModified)
   		return
   	}
		
	isContinue = true

	h = Conf.Headers
	h["Content-Type"] = mime.TypeByExtension(filepath.Ext(name))
	h["Last-Modified"] = t.UTC().Format(http.TimeFormat)
	if Conf.ETagSupport {
		h["ETag"] = i.ETag()	
	}
	return 
}

/**
 * Print json status
 */
func HttpPrintStatusJson(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", SERVER_SIGN)
	b := GetState().AsJson()
	w.Write(b)
}

/**
 * Print html status
 */
func HttpPrintStatusHtml(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Server", SERVER_SIGN)
	GetState().AsHtml(w)
}

