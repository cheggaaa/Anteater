package http

import (
	"cnst"
	"storage"
	"config"
	"net/http"
	"log"
	"fmt"
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
	s *storage.Storage
	conf *config.Config
}


func RunServer(s *storage.Storage) (server *Server) {
	server = &Server{
		s : s,
		conf : s.Conf,
	}
	server.Run()
	return
}

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


func (s *Server) ReadOnly(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: Log here
        	Err(500, w)
        }
	}()
	w.WriteHeader(200)
	w.Write([]byte("Read"))
}

func (s *Server) ReadWrite(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: Log here
        	Err(500, w)
        }
	}()
	s.Get("222", w)
}

func (s *Server) Get(name string, w http.ResponseWriter) {
	panic("ololo")
}


func Err(code int, w http.ResponseWriter) {
	w.Header().Add("Server", cnst.SIGN)
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf(ERROR_PAGE, httpErrors[code], httpErrors[code])))
}


