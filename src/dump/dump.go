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
)


func DumpTo(filename string, d interface{}) (n int, err error) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err = enc.Encode(d)
	if err != nil {
		return
	}
	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)	
	defer fh.Close()
	if err != nil {
		return 0, err
	}
	err = fh.Truncate(0)
	if err != nil {
		return 0, err
	}
	bytes := b.Bytes()
	n, err = fh.Write(bytes)
	return
}

func LoadData(filename string, d interface{}) (err error, exists bool) {	
	fh, err := os.Open(filename)
    if err != nil {
    	return
    }
    exists = true
    dec := gob.NewDecoder(fh)
    err = dec.Decode(d)
    return
}


