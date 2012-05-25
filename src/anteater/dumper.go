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

package anteater

import (
	"encoding/gob"
	"os"
	"bytes"
	"time"
)

type Data struct {
	ContainerLastId int32
	Containers []*ContainerDumpData
	Index      map[string]*FileInfo
}

func DumpData(filename string) error {
	Log.Debugln("Dump. Start dump data...")
	
	// Timer
	tm := time.Now()
	
	cs := []*ContainerDumpData{}
	for _, c := range(FileContainers) {
		cs = append(cs, c.GetDumpData())
	}
	IndexLock.Lock()
	d := &Data{ContainerLastId, cs, Index}
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	Log.Debugln("Dump. Encoder created... Start encode")
	err := enc.Encode(d)
	IndexLock.Unlock()
	if err != nil {
		return err
	}
	
	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	defer fh.Close()
	if err != nil {
		return err
	}
	Log.Debugln("Dump. Write to", filename)
	bytes := b.Bytes()
	n, err := fh.Write(bytes)
	if err != nil {
		return err
	}
	LastDump = time.Now()	
	IndexFileSize = int64(len(bytes))
	LastDumpTime = 	time.Now().Sub(tm)
	Log.Debugf("Dump. %d bytes successfully written to file\n", n)
	return nil
}

func LoadData(filename string) error {
	Log.Debugln("Try load index from", filename);
	
	fh, err := os.Open(filename)
    if err != nil {
    	return err
    }
    d := Data{}
    dec := gob.NewDecoder(fh)
    err = dec.Decode(&d)
    if err != nil {
    	return err
    }
    
    ContainerLastId = d.ContainerLastId
    
    for _, cd := range(d.Containers) {
    	c, err := ContainerFromData(cd)
    	if err != nil {
    		return err
    	}
    	FileContainers[c.Id] = c
    }
    
    Index = d.Index
    Log.Debugln("Index loaded")
    return nil
}
