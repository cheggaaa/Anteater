package main

import (
	"net/http"
	"time"
	"fmt"
	"io"
	"crypto/md5"
	"crypto/rand"
	"hash"
	mrand "math/rand"
	"os"
	"os/signal"
	"syscall"
	"errors"
	"utils"
	"flag"
)

var (
	ServerUrl string
)

func init() {
	mrand.Seed(time.Now().UnixNano() + time.Now().Unix())
	flag.StringVar(&ServerUrl, "s", "http://localhost:8081", "write server addr")
	flag.Parse()
}

func main() {

	clients := make([]*Client, 3)
	clients[0] = NewClient("s", 1, 1000000, 20,     1000)
	clients[1] = NewClient("s", 1, 1000000, 20,     1000)
	clients[2] = NewClient("s", 1, 1000000, 20,     10000)
	go clients[0].Run()
	go clients[1].Run()
	go clients[2].Run()
	
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGKILL, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	fmt.Printf("Catched signal %v\n", sig)
	clients[0].Stop()
	fmt.Printf("%#v\n", clients[0].Result)
	clients[1].Stop()
	fmt.Printf("%#v\n", clients[1].Result)
	clients[2].Stop()
	fmt.Printf("%#v\n", clients[2].Result)
}

type Client struct {
	c *http.Client
	prefix  string
	startId int64
	endId   int64
	minSize int64
	maxSize int64
	mode    int
	Result *Result
	list    map[int64]bool
	stop    bool
	stopped bool
}

type Result struct {
	Put, Post, Get, Delete, Link int64
	InBytes, OutBytes int64
	Error int64
	Time time.Duration
}



func NewClient(prefix string, startId, endId, minSize, maxSize int64) *Client {
	return &Client{
		c       : new(http.Client),
		prefix  : prefix,
		minSize : minSize,
		maxSize : maxSize,
		startId : startId,
		endId   : endId,
		Result  : &Result{},
		list    : make(map[int64]bool),
	}
}

func (c *Client) Run() {
	for  {
		if c.stop {
			c.stopped = true
			break
		}
		c.run()
	}
}

func (c *Client) run() {
	id := c.putId()
	s, err := c.upload(id)
	c.list[id] = true
	if err != nil {
		fmt.Printf("[%d] upload err: %v\n", id, err)
		c.Result.Error++
	} else {
		fmt.Printf("[%d] uploaded %s\n", id, utils.HumanBytes(s))
		c.Result.Put++
	}
}

func (c *Client) Stop() {
	return
	c.stop = true
	for {
		time.Sleep(time.Millisecond * 10)
		if c.stopped {
			break
		}
	}
	for id, r := range c.list {
		if r {
			err := c.fdelete(id)
			if err != nil {
				fmt.Println("Delete err:", id, err)
				c.Result.Error++
			} else {
				fmt.Println("Deleted", id)
				c.Result.Put++
			}
		}
	}
}

func (c *Client) upload(id int64) (size int64, err error) {
	purl := c.url(id)
	reader := c.reader()
	req, err := http.NewRequest("PUT", purl, reader)
	if err != nil {
		return
	}
	size = reader.Size
	req.Header.Set("Content-Type", "application/octed-stream")
	req.ContentLength = size
	defer req.Body.Close()
	res, err := c.c.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	c.Result.OutBytes += size
	switch res.StatusCode {
		case 200, 201:
			if reader.Md5() != res.Header.Get("X-Ae-Md5") {
				return 0, errors.New("Md5 mismatched! " + purl)
			}
		default:
			return 0, errors.New(res.Status)
	}
	return
}

func (c *Client) fdelete(id int64) (err error) {
	purl := c.url(id)
	req, err := http.NewRequest("DELETE", purl, nil)
	if err != nil {
		return
	}
	res, err := c.c.Do(req)
	if err != nil {
		return
	}
	switch res.StatusCode {
		case 200, 201, 204:
			break
		default:
			return errors.New(res.Status)
	}
	return
}

func (c *Client) putId() (id int64) {
	return mrand.Int63n(c.endId - c.startId) + c.startId
}


func (c *Client) url(id int64) string {
	return fmt.Sprintf("%s/%s%d.bin", ServerUrl, c.prefix, id)
}

func (c *Client) reader() *LimitMd5Reader {
	n := mrand.Int63n(c.maxSize - c.minSize) + c.minSize
	return &LimitMd5Reader{io.LimitReader(rand.Reader, n), md5.New(), n}
}

// Limit md5 reader
type LimitMd5Reader struct {
	io.Reader
	h hash.Hash
	Size int64
}

func (r *LimitMd5Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil {
		return
	}
	r.h.Write(p[:n])
	return
}
func (r *LimitMd5Reader) Md5() string {
	return fmt.Sprintf("%x", r.h.Sum(nil))
}
