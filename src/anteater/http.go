package anteater

import (
	"net/http"
	"fmt"
	"strconv"
	"io"
	"sort"
	"strings"
	"mime"
	"time"
)

const (
	version   = "0.004"
	errorPage = "<html><head><title>%s</title></head><body><center><h1>%s</h1></center><hr><center>Anteater " + version + "</center></body></html>\n"
)


var httpErrors map[int]string = map[int]string{
	400: "Invalid request",
	404: "404 Not Found",
	405: "405 Method Not Allowed",
	411: "411 Length Required",
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

func HttpRead(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[1:]
	if len(filename) == 0 {
		errorFunc(w, 404)
		return
	}
	if r.Method == "GET" {
		getFile(filename, w, r)
		return
	}
	errorFunc(w, 501)
}

func HttpReadWrite(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[1:]
	if len(filename) == 0 {
		errorFunc(w, 404)
		return
	}
	
	if filename == "status" {
		printHtml(w)
		return
	}

	switch r.Method {
	case "GET":
		getFile(filename, w, r)
		return
	case "POST":
		saveFile(filename, w, r)
		return
	case "PUT":
		deleteFile(filename)
		saveFile(filename, w, r)
		return
	case "DELETE":
		st := deleteFile(filename)
		if !st {
			errorFunc(w, 404)
		}
		return
	default:
		Log.Infoln("Unhandled method", r.Method)
		errorFunc(w, 501)
	}
}

func errorFunc(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if status == 405 {
		w.Header().Set("Allow", "GET")
	}
	fmt.Fprintf(w, errorPage, httpErrors[status], httpErrors[status])
}


func getFile(name string, w http.ResponseWriter, r *http.Request) {
	i, ok := IndexGet(name)
	if !ok {
		errorFunc(w, 404)
		return
	}

	h, isContinue := httpHeadersHandle(name, i, w, r)
	
	if  ! isContinue {
		return
	}
	
	reader := i.GetReader()
	
	for k, v := range(h) {
		w.Header().Set(k, v)
	}
	
	http.ServeContent(w, r, name, time.Unix(i.T, 0), reader)
}

func saveFile(name string, w http.ResponseWriter, r *http.Request) {
	_, ok := IndexGet(name)
	if ok {
		errorFunc(w, 405)
		return
	}

	file := r.Body
	size := r.ContentLength
	
	if size == 0 {
		 errorFunc(w,  411)
		 return
	}
	
	Log.Debugln("Start upload file", name, size, "bytes")
	f, fi, err := GetFile(name, size)
	
	var written int64
	for {
		buf := make([]byte, 1024*1024)
		nr, er := io.ReadAtLeast(file, buf, 1024*1024)
		if nr > 0 {
			nw, ew := f.WriteAt(buf[0:nr], written)
			if nw > 0 {
				written += int64(nw)
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
			errorFunc(w, 500)
			return
		}
	}
	
	Log.Debugf("File %s (%d:%d) uploaded.\n", name, fi.ContainerId, fi.Id)	
	fmt.Fprintf(w, "OK\nSize:%d\nETag:%s\n", size, fi.ETag())	
}

func deleteFile(name string) bool {
	i, ok := IndexGet(name)
	if !ok {
		return false
	}
	delete(Index, name)
	FileContainers[i.ContainerId].Delete(i.Id, i.Start, i.Size)
	return true
}


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
	isContinue = true	
	h = Conf.Headers
	
	h["Content-Type"] = getType(name)
	h["Content-Length"] = strconv.FormatInt(i.Size, 10)
	if Conf.ETagSupport {
		h["ETag"] = i.ETag()	
	}
	return
}

func printHtml(w http.ResponseWriter) {
	GetFileLock.Lock()
	defer GetFileLock.Unlock()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var ids []int = make([]int, 0)
	for _, cn := range(FileContainers) {
		ids = append(ids, cn.Id)
	}
	sort.Ints(ids)
	var cnt *Container
	var html string = ""
	for _, id := range(ids) {
		cnt = FileContainers[id]
		html += "<div style=\"float:left;margin:20px;\">"
		html += "<h3>C " + strconv.FormatInt(int64(cnt.Id), 10) + "</h3>"
		html += "<ul>"
		html += "<li>File count: <b>" + strconv.FormatInt(cnt.Count, 10) + "</b></li>"
		html += "<li>LastId: <b>" + strconv.FormatInt(cnt.LastId, 10) + "</b></li>"
		html += "<li>Size: <b>" + strconv.FormatInt(cnt.Size, 10) + "</b></li>"
		html += "<li>Offset: <b>" + strconv.FormatInt(cnt.Offset, 10) + "</b></li>"
		html += "<li>Max Space Size: <b>" + strconv.FormatInt(cnt.MaxSpace(), 10) + "</b></li>"
		html += "<li>Max Spaces Size: <b>" + strconv.FormatInt(cnt.MaxSpaceSize, 10) + "</b></li>"
		html += "</ul>"
		html += "<h4>Spaces</h4>"
		html += cnt.Spaces.ToHtml(10, cnt.Size)
		html += "</div>"
	}
	fmt.Fprintf(w, "<html><body><div>%s</div></body></html>", html)
}

func getType(name string) string {
	t := strings.Split(name, ".")
	var ext string = "." + t[len(t) - 1]
	fType := mime.TypeByExtension(ext)
	if len(fType) == 0 {
		fType = "application/octed-stream"
	}
	return fType
}

