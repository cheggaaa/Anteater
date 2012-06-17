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
	"fmt"
	"io"
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
	size, err := dumpTo(filename, d)
	IndexLock.Unlock()
	if err != nil {
		return err
	}
	IndexFileSize = int64(size)
	LastDump = time.Now()	
	LastDumpTime = 	time.Now().Sub(tm)
	Log.Debugf("Dump. %d bytes successfully written to file\n", IndexFileSize)
	return nil
}

func dumpTo(filename string, d *Data) (int, error) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	Log.Debugln("Dump. Encoder created... Start encode")
	err := enc.Encode(d)
	if err != nil {
		return 0, err
	}
	
	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	defer fh.Close()
	if err != nil {
		return 0, err
	}
	Log.Debugln("Dump. Write to", filename)
	bytes := b.Bytes()
	n, err := fh.Write(bytes)
	if err != nil {
		return 0, err
	}
	return n, nil
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


func DumpAllTo(path string) error {
	Log.Debugln("Strat dump all")
	var err error
	
	err = os.MkdirAll(path, 0666)
	
	if err != nil {
		return err
	}
		
	cs := []*ContainerDumpData{}
	var dumpIndex map[string]*FileInfo = make(map[string]*FileInfo)
	
	for _, c := range FileContainers {		
		cs = append(cs, c.GetDumpData())
		dumpIndex, err = dumpContainer(path, c, dumpIndex)
		if err != nil {
			Log.Warn(err)
			return err
		}
	}
	
	d := &Data{ContainerLastId, cs, dumpIndex}
	filename := fmt.Sprintf("%s/file.index", path)
	_, err = dumpTo(filename, d)
	if err != nil {
		Log.Warn(err)
		return err
	}
	Log.Debugln("All containers dumped");
	return nil
}

func dumpContainer(path string, c *Container, i map[string]*FileInfo) (map[string]*FileInfo, error) {
	Log.Debugln("Strat dump container", c.Id);
	// copy container file 
	c.Disable()
	c.WLock.Lock()
	defer c.WLock.Unlock()
	defer c.Enable()
	
	dpath := fmt.Sprintf("%s/file.data.%d", path, c.Id)
	dst, err := os.OpenFile(dpath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer dst.Close()
	dst.Truncate(c.Size)
	reader := io.NewSectionReader(c.F, 0, c.Size)
	n, err := io.Copy(dst, reader)
	Log.Debugln("Done!", n, "bytes copied");
	if err != nil {
	   return nil, err
	}
	
	//copy index
	IndexLock.Lock()
	defer IndexLock.Unlock()
	for n, fi := range Index {
		if fi.ContainerId == c.Id {
			i[n] = fi
		}	
	}
	
	return i, nil
}
