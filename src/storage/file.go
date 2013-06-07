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

package storage

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
	//"aelog"
	"fmt"
)

type File struct {
	Hole  // inherits
	Name  string
	Md5   []byte
	FSize int64
	Time  time.Time

	c         *Container
	ctype     string
	deleted   bool
	openCount int32
}

func (f *File) Init(c *Container) {
	f.c = c
}

// implement Space interafce
func (f *File) IsFree() bool {
	return false
}

// mark file as Open
func (f *File) Open() (err error) {
	if f.deleted {
		return errors.New("File deleted")
	}
	atomic.AddInt32(&f.openCount, 1)
	return
}

// need call after open
func (f *File) Close() {
	if atomic.AddInt32(&f.openCount, -1) == 0 && f.deleted {
		f.Delete()
	}
}

// mark as deleted
func (f *File) Delete() {
	f.deleted = true
	if atomic.LoadInt32(&f.openCount) == 0 && f.c != nil {
		f.c.Delete(f)
	}
}

// return io.Reader
func (f *File) GetReader() *Reader {
	return newReader(f.c.f, f.Off, f.FSize, f.c.s)
}

func (f *File) WriteAt(b []byte, off int64) (int, error) {
	if off+int64(len(b)) > f.FSize {
		panic("Can't write. Overflow allocated size")
	}
	off = off + f.Off
	return f.c.f.WriteAt(b, off)
}

// return http E-Tag
func (f *File) ETag() string {
	return strconv.FormatInt(int64(f.c.Id), 36) + "." + strconv.FormatInt(f.Time.UnixNano(), 36)
}

// Return content type file or application/octed-stream if can't
func (f *File) ContentType() (ctype string) {
	ctype = f.ctype
	if ctype == "" {
		ctype = mime.TypeByExtension(filepath.Ext(f.Name))
		if ctype == "" {
			var buf [512]byte
			n, _ := io.ReadFull(f.GetReader(), buf[:])
			b := buf[:n]
			ctype = http.DetectContentType(b)
		}
		f.ctype = ctype
	}
	return
}

// copy content from io.Reader
func (f *File) ReadFrom(r io.Reader) (written int64, err error) {
	h := md5.New()
	var bs int
	if f.FSize > 128*1024 {
		bs = 64 * 1024
	} else {
		bs = int(f.FSize)
	}
	buf := make([]byte, bs)
	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			nw, ew := f.WriteAt(buf[0:nr], written)
			if nw > 0 {
				written += int64(nw)
				h.Write(buf[0:nw])
				f.c.s.Stats.Traffic.Input.AddN(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	f.Md5 = h.Sum(nil)
	return
}

// string md5
func (f *File) Md5S() string {
	return hex.EncodeToString(f.Md5)
}

func (f *File) CheckMd5() (err error) {
	if err = f.Open(); err != nil {
		return
	}
	defer f.Close()
	r := f.GetReader()
	h := md5.New()
	io.Copy(h, r)
	if hex.EncodeToString(h.Sum(nil)) != f.Md5S() {
		err = fmt.Errorf("File %s. MD5 mismatched: %s vs %s", f.Name, hex.EncodeToString(h.Sum(nil)), f.Md5S())
	}
	return
}
