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
	"fmt"
	"io"
)

type FileInfo struct {
	Id          int64
	ContainerId int32
	Start       int64
	Size        int64
	Md5         []byte
	C           int64
	T			int64
}



func init() {
	IndexLock.Lock()
	defer IndexLock.Unlock()
	Index = make(map[string]*FileInfo, 1000)
}

func IndexSet(name string, f *FileInfo) {
	IndexLock.Lock()
	defer IndexLock.Unlock()
	Index[name] = f
	return
}


func IndexGet(name string) (i *FileInfo, ok bool) {
	i, ok = Index[name]
	return 
}

func IndexDelete(name string) (*FileInfo, bool) {
	IndexLock.Lock()
	defer IndexLock.Unlock()
	if Index[name] == nil {
		return nil, false
	} 
	i := Index[name]
	delete(Index, name)
	return i, true
}

func (f *FileInfo) GetReader() *io.SectionReader {
	c := FileContainers[f.ContainerId]
	return io.NewSectionReader(c.F, f.Start,f.Size)
}

func (f *FileInfo) ETag () string {
	return fmt.Sprintf("\"%x%x%x\"", f.ContainerId, f.T, f.Id)
}
