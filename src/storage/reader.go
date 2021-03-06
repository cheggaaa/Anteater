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
	"errors"
	"io"
)

func newReader(r io.ReaderAt, off int64, n int64, s *Storage) *Reader {
	return &Reader{r, off, off, off + n, s}
}

type Reader struct {
	r                io.ReaderAt
	base, off, limit int64
	s                *Storage
}

func (s *Reader) Read(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, io.EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.ReadAt(p, s.off)
	s.s.Stats.Traffic.Output.AddN(n)
	s.off += int64(n)
	return
}

func (s *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	default:
		return 0, errors.New("Seek: invalid whence")
	case 0:
		offset += s.base
	case 1:
		offset += s.off
	case 2:
		offset += s.limit
	}
	if offset < s.base || offset > s.limit {
		return 0, errors.New("Seek: invalid offset")
	}
	s.off = offset
	return offset - s.base, nil
}

func (s *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit-s.base {
		return 0, io.EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.ReadAt(p, off)
	s.s.Stats.Traffic.Output.AddN(n)
	return
}

func (s *Reader) WriteTo(dst io.Writer) (written int64, err error) {
	var bs int64
	if sz := s.Size(); sz > 100*1024 {
		bs = 64 * 1024
	} else {
		bs = sz
	}
	buf := make([]byte, bs)
	for {
		nr, er := s.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
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
	return written, err
}

func (s *Reader) Size() int64 { return s.limit - s.base }
