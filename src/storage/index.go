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
	"sync"
	"sync/atomic"
	"errors"
)

var ErrIndexAlredyExists = errors.New("Index already exists")

type Index struct {
	Files map[string]*File
	m *sync.Mutex
	v uint64
}

type IndexDump struct {
	F map[string]FileDump
	V uint64
}

func (i *Index) Add(name string, file *File) (err error) {
	i.m.Lock()
	defer i.m.Unlock()
	_, isset := i.Files[name]
	if isset {
		err = ErrIndexAlredyExists
		return
	}
	i.Files[name] = file
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

func (i *Index) DumpData() (dump IndexDump) {
	df := make(map[string]FileDump, len(i.Files)) 
	for n, f := range i.Files {
		df[n] = f.DumpData()
	}
	dump.F = df
	dump.V = i.v
	return
}

func (i *Index) Version() uint64 {
	return atomic.LoadUint64(&i.v)
}

func (id *IndexDump) Restore(s *Storage) *Index {
	files := make(map[string]*File, len(id.F))
	for n, f := range id.F {
		files[n] = &File{
			CId : f.CId,
			Md5 : f.Md5,
			Size : f.Size,
			Start : f.Start,
			Time : f.Time,			
		}
		files[n].Init(s, n)
	} 
	return &Index{files, &sync.Mutex{}, id.V}
}