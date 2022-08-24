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
)


func DumpTo(filename string, d interface{}) (n int64, err error) {
	tmpfile := filename + ".tmp"
	tf, err := os.OpenFile(tmpfile, os.O_CREATE|os.O_RDWR, 0666)		
	if err != nil {
		return
	}
	tf.Truncate(0)
	defer tf.Close()
	
	enc := gob.NewEncoder(tf)
	err = enc.Encode(d)
	if err != nil {
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

func LoadData(filename string, d interface{}) (err error, exists bool) {	
	fh, err := os.Open(filename)
    if err != nil {
    	return
    }
    defer fh.Close()
    exists = true
	dec := gob.NewDecoder(fh)
    err = dec.Decode(d)
    return
}


