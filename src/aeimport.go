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

package main

import (
	"io"
	"os"
	"strings"
	"fmt"
	"flag"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"errors"
	"crypto/md5"
	"bufio"
)

const (
	DEFAULT_HOST = "127.0.0.1"
	DEFAULT_PORT = "8081"
)

const USAGE = `
Usage:
	aeimport [-s="host:port"] [-v] [-m=method] [-c=concurrency] [-p=prefix] /path/to/dir
		OR
	aeimport [-s="host:port"] [-v] [-m=method] [-c=concurrency] [-p=prefix] /path/to/file_with_urls
	
Options:
	-s=server_addr
		default addr: 127.0.0.1:8081
		examples:
			-s="192.168.1.2" will be 192.168.1.2:8081
			-s=":82" will be 127.0.0.1:82
			-s="anteater.local:82" will be anteater.local:82
		
	-m=method
		upload method
		must be PUT or POST (by default POST)
		
	-c=concurrecy
		concurrecy level, by default 1
		
	-p=prefix
		prefix for upload path
		for example if file path = 2003/123.jpg and prefix = photo
		end path will be photo/2003/123.jpg
	
	-v 
		verbose - show additional info	
`;

var (
	ShowHelp bool
	Dir string
	Method string
	Concurrency int
	Verbose bool
	UploadUrl string
	Paths []string = make([]string, 0)
	PathC int64 = -1
	FilesSize int64 = 1
	Prefix string
	UrlReader *bufio.Reader
	ErrNotDir = errors.New("Not dir")
	ModeUrls bool
)

func init() {
	flag.BoolVar(&Verbose, "v", false, "Verbose")
	flag.IntVar(&Concurrency, "c", 1, "Concurrency level")
	flag.StringVar(&Method, "m", "POST", "Upload method POST or PUT")
	flag.StringVar(&Prefix, "p", "", "Prefix")
	ParseArgs()
}


func ParseArgs() {
	ServerAddr := flag.String("s", DEFAULT_HOST + ":" + DEFAULT_PORT, "Server addr")
	
	flag.Parse();
	var s int = 1
	s += flag.NFlag()

	if len(os.Args) - s >= 1 {
		Dir = os.Args[s]
		s++
	} else {
		ShowHelp = true
		return
	}
	
	// check server addr
	*ServerAddr = strings.Trim(*ServerAddr, "/")
	addr := strings.Split(*ServerAddr, ":")
	var host, port string
	if len(addr) == 1 {
		host = addr[0]
		port = DEFAULT_PORT
	} else if len(addr) == 2 {
		host = addr[0]
		port = addr[1]
	}
	if len(host) == 0 {
		host = DEFAULT_HOST 
	}
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	UploadUrl = "http://" + host + ":" + port
	if Prefix != "" {
		Prefix = strings.Trim(Prefix, "/") + "/"
	}

	return
}



func main() {
	if ShowHelp {
		fmt.Println(USAGE)
		return
	}
	
	switch strings.ToLower(Method) {
		case "post":
			Method = "POST"
			break
		case "put":
			Method = "PUT"
			break
		default:
			fmt.Printf("Unknown method %s. Must be POST or PUT\n", Method)
			return
	}
	
	if Concurrency < 1 {
		fmt.Printf("Conccurency can't be %d. Will 1\n", Concurrency)
		Concurrency = 1
	}
	os.Chdir(Dir)
	Log("Start scan directory")
	err := Scan(Dir, true)
	
	if err == ErrNotDir {
		err = nil
		ModeUrls = true
		err = ScanFile(Dir)
	} 
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Found %d files. %.2f Mb\n", len(Paths), float64(FilesSize) / 1024 / 1024)
	
	wg := new(sync.WaitGroup)
	wg.Add(Concurrency)
	for i := 0; i < Concurrency; i++ {
		LogD("Start client", i + 1)
		go RunClient(wg, i)
	}
	wg.Wait()
}


func Scan(p string, first bool) (err error) {
	
	f, err := os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()
	
	s, err := f.Stat()
	if err != nil {
		return
	}
	if ! s.IsDir() {
		err = ErrNotDir
		return
	}
	
	files, err := f.Readdir(0)
	if err != nil {
		return
	}
	
	if first {
		p = ""
	}
	
	if p != "" {
		p = p + string(os.PathSeparator)
	}
	
	for _, fi := range files {
		path := p + fi.Name()
		if fi.IsDir() {
			Scan(path, false)
		} else {
			FilesSize += fi.Size()
			Paths = append(Paths, path)
		}
	}

	return
}


func ScanFile(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	UrlReader = bufio.NewReader(f)
	return
}

func RunClient(wg *sync.WaitGroup, cl int) {
	c := new(http.Client)
	for {
		var path string
		var end bool
		if ModeUrls {
			path, end = GetNextUrl()
		} else {
			path, end = GetNextPath()
		}
		if end {
			break
		}
		if path == "" {
			continue
		}
		err := Upload(c, path, cl)
		if err != nil {
			Log("Error upload", path, err, "\n")
		}
	}
	wg.Done()
}

func GetNextPath() (string, bool) {
	i := atomic.AddInt64(&PathC, 1)
	if (i + 1) > int64(len(Paths)) {
		return "", true
	}
	path := Paths[i]
	return path, false
}

var UrlMutex = &sync.Mutex{}
var UrlsEnd bool

func GetNextUrl() (string, bool) {
	UrlMutex.Lock()
	defer UrlMutex.Unlock()
	if UrlsEnd {
		return "", true
	}
	
	line, isPrefix, err := UrlReader.ReadLine()
    if ! isPrefix && err == nil {
    	return string(line), false
    }
    UrlsEnd = true
    return "", false
}


func Upload(client *http.Client, path string, cl int) error  {
	if ModeUrls {
		purl, err := url.Parse(path)
		if err != nil {
			return err
		}
		furl := UploadUrl + "/" + Prefix + strings.TrimLeft(purl.Path, "/") + "?url=" + url.QueryEscape(path)
		LogD(cl,": Upload to", furl)
		req, err := http.NewRequest(Method, furl, nil)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		switch res.StatusCode {
			case 200, 201:
				return nil
			default:
				return errors.New(res.Status)
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	s, err := f.Stat()
	if err != nil {
		return err
	}
	
	furl := UploadUrl + "/" + Prefix + path
	
	LogD(cl,": Upload to", furl)
	
	h := md5.New()
    io.Copy(h, f)
    fMd5 := fmt.Sprintf("%x", h.Sum(nil)) 
	
	req, err := http.NewRequest(Method, furl, f)
	defer req.Body.Close()
	t := make([]byte, 512)
	io.ReadFull(f, t)
	f.Seek(0, 0)
	req.Header.Set("Content-Type", http.DetectContentType(t))
	req.ContentLength = s.Size()
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	
	switch res.StatusCode {
		case 200, 201:
			if fMd5 != res.Header.Get("X-Ae-Md5") {
				return errors.New("Md5 mismatched! " + path)
			}
		default:
			return errors.New(res.Status)
	}
	
	return nil
}


func LogD(a...interface{}) {
	if Verbose {
		Log(a...)
	}
}

func Log(a...interface{}) {
	fmt.Println(a...)
}

