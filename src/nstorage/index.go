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

package nstorage

import (
	"errors"
	"sync"
	"sync/atomic"
)

type Index struct {
	Files map[string]*File
	m     *sync.Mutex
	v     uint64
}

// Add new file to index
func (i *Index) Add(file *File) (err error) {
	i.m.Lock()
	defer i.m.Unlock()
	if _, exists := i.Files[file.Name]; exists {
		err = errors.New("Index already exists: " + file.Name)
		return
	}
	i.Files[file.Name] = file
	atomic.AddUint64(&i.v, 1)
	return
}

func (i *Index) Get(name string) (f *File, ok bool) {
	f, ok = i.Files[name]
	return
}

func (i *Index) Delete(name string) (f *File, ok bool) {
	i.m.Lock()
	defer i.m.Unlock()
	f, ok = i.Files[name]
	if ok {
		delete(i.Files, name)
		atomic.AddUint64(&i.v, 1)
	}
	return
}

func (i *Index) Version() uint64 {
	return atomic.LoadUint64(&i.v)
}
