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

package dump

import (
	"bufio"
	"io"
	"os"
)

type Writer interface {
	Write(wr io.Writer) error
}

func DumpTo(filename string, ewr Writer) (n int64, err error) {
	tmpfile := filename + ".tmp"
	tf, err := os.OpenFile(tmpfile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return
	}
	tf.Truncate(0)
	defer tf.Close()
	wr := bufio.NewWriterSize(tf, 16*1024)
	for {
		e := ewr.Write(wr)
		if e != nil {
			if e == io.EOF {
				break
			} else {
				return 0, e
			}
		}
	}
	if err = wr.Flush(); err != nil {
		return
	}
	os.Remove(filename)
	err = os.Rename(tmpfile, filename)
	if err != nil {
		return
	}
	i, _ := tf.Stat()
	n = i.Size()
	return
}

type ResultReader struct {
	B *bufio.Reader
	f io.Closer
}

func (rr *ResultReader) Close() error {
	return rr.f.Close()
}

func LoadData(filename string) (rr *ResultReader, err error, exists bool) {
	fh, err := os.Open(filename)
	if err != nil {
		return
	}
	exists = true
	rr = &ResultReader{B: bufio.NewReader(fh), f: fh}
	return
}
