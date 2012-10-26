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
	"encoding/gob"
	"os"
	"bytes"
	"compress/gzip"
	"io"
	"errors"
	"fmt"
)


func DumpTo(filename string, d interface{}) (n int, err error) {
	b := new(bytes.Buffer)
	zb := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err = enc.Encode(d)
	if err != nil {
		return
	}
	dumpfile := filename + ".td"
	fh, err := os.OpenFile(dumpfile, os.O_CREATE|os.O_RDWR, 0666)		
	if err != nil {
		return
	}
	defer fh.Close()

	err = fh.Truncate(0)
	if err != nil {
		return
	}
	
	w, err := gzip.NewWriterLevel(zb, gzip.BestSpeed)
	if err != nil {
		return
	}
	
	n, err = w.Write(b.Bytes())
	if err != nil {
		return
	}
	err = w.Close()
	if err != nil {
		return
	}
	n, err = fh.Write(zb.Bytes())
	if err != nil {
		return
	}
	
	// Copy to destination file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)	
	if err != nil {
		return
	}
	defer f.Close()
	f.Truncate(0)
	fh.Seek(0, 0)
	cp, err := io.Copy(f, fh)
	if err != nil {
		return
	}
	if cp != int64(n) {
		err = errors.New(fmt.Sprintf("Writed %d bytes, but copied only %d", n, cp))
		return 
	}
	return
}

func LoadData(filename string, d interface{}) (err error, exists bool) {	
	fh, err := os.Open(filename)
    if err != nil {
    	return
    }
    defer fh.Close()
    exists = true
	
	r, err := gzip.NewReader(fh)
	if err != nil {
		return
	}
	defer r.Close()
	
	dec := gob.NewDecoder(r)
    err = dec.Decode(d)		
     
    return
}


