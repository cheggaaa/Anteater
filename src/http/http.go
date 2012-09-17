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
)

const (
	ERROR_PAGE = "<html><head><title>%s</title></head><body><center><h1>%s</h1></center><hr><center>" + cnst.SIGN + "</center></body></html>\n"
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

type Server struct {
	stor *storage.Storage
	conf *config.Config
}

// Create new server and run it
func RunServer(s *storage.Storage) (server *Server) {
	server = &Server{
		stor : s,
		conf : s.Conf,
	}
	server.Run()
	return
}

// Run all servers
func (s *Server) Run() {
	run := func(handler http.Handler, addr string) { 
		s := &http.Server{
			Addr:         addr,
			Handler:      handler,
		}
		log.Fatal(s.ListenAndServe())
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
			// TODO: Log here
			fmt.Println(rec)
        	Err(500, w)
        }
        r.Body.Close()
	}()
	filename := Filename(r)
	if len(filename) == 0 {
		Err(404, w)
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
	}
	Err(501, w)
}

// Http handler for read-write server
func (s *Server) ReadWrite(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", cnst.SIGN)
	defer func() {
		if rec := recover(); rec != nil {
			// TODO: Log here
			fmt.Println(rec)
        	Err(500, w)
        }
        r.Body.Close()
	}()
	filename := Filename(r)
	switch filename {
		case "":
			Err(404, w)
			fmt.Println("Empty filenamr")
			return;
		case s.conf.StatusHtml:
			fmt.Println("StatusHtml")
			Err(500, w)
			return
		case s.conf.StatusJson:
			fmt.Println("StatusJson")
			s.StatsJson(w)
			return
	}
	
	m := r.Header.Get("X-Http-Method-Override")
	if m == "" {
		m = r.Method
	}
	
	switch m {
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
		s.Delete(filename, nil)
		s.Save(filename, w, r)
		return
	case "DELETE":
		s.Delete(filename, w)
		return
	default:
		Err(501, w)
	}
}

func (s *Server) Get(name string, w http.ResponseWriter, r *http.Request, writeBody bool) {
	f, ok := s.stor.Get(name)
	if ! ok {
		Err(404, w)
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
		return
	}
	
	reader := f.GetReader()
	
	if f.Size > s.conf.ContentRange {
		http.ServeContent(w, r, name, f.Time, reader)
	} else {
		io.Copy(w, reader)
	}
	
}

func (s *Server) Save(name string, w http.ResponseWriter, r *http.Request) {
	_, ok := s.stor.Get(name)
	if ok {
		// File exists
		Err(409, w)
		return
	}
	
	reader := r.Body
	size := r.ContentLength
	
	if size <= 0 {
		Err(411, w)
		return
	}
	
	if size > s.conf.ContainerSize {
		Err(413, w)
		return
	}
	
	s.stor.Stats.Counters.Add.Add()
	
	f := s.stor.Add(name, reader, size)
	
	w.Header().Set("X-Ae-Md5", fmt.Sprintf("%x", f.Md5));	
	w.Header().Set("Etag", f.ETag());
	w.Header().Set("Location", name);
	w.WriteHeader(http.StatusCreated)
}


func (s *Server) Delete(name string, w http.ResponseWriter) {
	ok := s.stor.Delete(name)
	if ok {
		if w != nil {
			w.WriteHeader(http.StatusNoContent)
		}
		s.stor.Stats.Counters.Delete.Add()
		return
	} else {
		if w != nil {
			Err(404, w)
		}
	}
}

func (s *Server) StatsJson(w http.ResponseWriter) {
	b := s.stor.GetStats().AsJson()
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(b)
}


func Err(code int, w http.ResponseWriter) {
	body := []byte(fmt.Sprintf(ERROR_PAGE, httpErrors[code], httpErrors[code]))
	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	w.Header().Add("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(code)
	w.Write(body)
}


/**
 * Return filename without first slashes
 */
func Filename(r *http.Request) string {
	var i int
	for _, s := range r.URL.Path {	
		if string(s) != "/" {
			break
		}
		i++
	}
	return r.URL.Path[i:]
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

