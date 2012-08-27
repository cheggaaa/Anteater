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
	"errors"
	"io"
	"time"
	"crypto/md5"
	"fmt"
)


var Targets = []int{TARGET_SPACE_EQ, TARGET_SPACE_FREE, TARGET_NEW}

/**
 * Create and write file to storage
 */
func WriteFileToStorage(r io.ReadCloser, name string, size int64) (*FileInfo, error) {
	f, err := GetFile(name, size)
	fi := f.Info()
	
	isOk := false

	defer func() {
		r.Close()
		if isOk {
			IndexSet(name, fi)
			HttpCn.CAdd()
		} else {
			FileContainers[fi.ContainerId].Delete(fi)
		}
	}()
	
	var written int64
	h := md5.New()
	for {
		buf := make([]byte, 100*1024)
		nr, er := io.ReadFull(r, buf)
		if nr > 0 {
			nw, ew := f.WriteAt(buf[0:nr], written)
			if nw > 0 {
				written += int64(nw)
				h.Write(buf[0:nw])
			}
			if ew != nil {
				err = ew
				break
			}
		}
		if er != nil {
			err = er
			break
		}
		if err != nil {
			return nil, err
		}
	}
	
	fi.Md5 = h.Sum(nil)	
	isOk = true
	
	Log.Debugln("File saved. Name:", name, "Size:", size, "Md5:", fmt.Sprintf("%x", fi.Md5))
	
	return fi, nil
}

/**
 * Allocate space, select container and create file object
 */
func GetFile(name string, size int64) (*File, error) {
	GetFileLock.Lock()
	defer GetFileLock.Unlock()
	for _, target := range(Targets) {
		for _, c := range(Containers) {
			if c.MaxSpace() >= size {
				f, err := c.Allocate(size, target)
				if err == nil {
					AllocCn.CTarget(target)
					return f, nil
				}
			}
		}
	}
	cId, err := NewContainer(DataPath)
	if err != nil {
		return nil, err
	}
	 
	f, err := Containers[cId].Allocate(size, TARGET_NEW)
	if err != nil {
		return nil, err
	}
		
	return f, nil
}


type File struct {
	Id    int64
	C    *Container
	Start int64
	Size  int64
}

func (f *File) WriteAt(b []byte, off int64) (int, error) {
	if off + int64(len(b)) > f.Size {
		return 0, errors.New("Can't write. Overflow allocated size")		
	}
	off = off + f.Start
	return f.C.F.WriteAt(b, off)
}

func (f *File) GetReader() *io.SectionReader {
	return io.NewSectionReader(f.C.F, f.Start,f.Size)
}

func (f *File) Info() *FileInfo {
	return &FileInfo{f.Id, f.C.Id, f.Start, f.Size, nil, 0, time.Now().Unix()}
}

