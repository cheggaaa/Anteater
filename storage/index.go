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
	"errors"
	"fmt"
	"github.com/cheggaaa/Anteater/src/aelog"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	ErrConflict = errors.New("Conflict")
)

type Index struct {
	Root *Node
	m    *sync.Mutex
	v, c int64
}

func (i *Index) Init() {
	i.m = &sync.Mutex{}
	i.Root = &Node{}
}

// Add new file to index
func (i *Index) Add(file *File) (err error) {
	i.m.Lock()
	defer i.m.Unlock()
	return i.add(file)
}

func (i *Index) Get(name string) (f *File, ok bool) {
	i.m.Lock()
	defer i.m.Unlock()
	return i.get(name)
}

func (i *Index) Delete(name string) (f *File, ok bool) {
	i.m.Lock()
	defer i.m.Unlock()
	return i.delete(name)
}

func (i *Index) Rename(name, newName string) (f *File, err error) {
	i.m.Lock()
	defer i.m.Unlock()
	// check file
	f, ok := i.get(name)
	if !ok {
		err = ErrFileNotFound
		return
	}
	// check new name
	_, ok = i.get(newName)
	if ok {
		err = ErrConflict
		f = nil
		return
	}
	// rename
	f.Name = newName
	i.delete(name)
	if err = i.add(f); err != nil {
		// rollback
		f.Name = name
		i.add(f)
		err = fmt.Errorf("Can't rename %s to %s: %v", name, newName, err)
		f = nil
		return
	}
	return
}

func (i *Index) List(prefix string, maxnesting int) (names []string, err error) {
	parts := make([]string, 0)
	if prefix != "" {
		parts = i.explode(prefix)
	}
	return i.Root.List(parts, 0, maxnesting)
}

func (i *Index) Count() int64 {
	return atomic.LoadInt64(&i.c)
}

func (i *Index) Version() int64 {
	return atomic.LoadInt64(&i.v)
}

func (i *Index) get(name string) (f *File, ok bool) {
	parts := i.explode(name)
	var err error
	if f, err = i.Root.Get(parts, 0); err == nil {
		ok = true
		return
	}
	if err != ErrFileNotFound {
		aelog.Warnf("Error while get from index. %s: %v", name, err)
	}
	return
}

func (i *Index) delete(name string) (f *File, ok bool) {
	parts := i.explode(name)
	var err error
	if f, err = i.Root.Delete(parts, 0); err == nil {
		ok = true
		atomic.AddInt64(&i.v, 1)
		atomic.AddInt64(&i.c, -1)
		return
	}
	if err != ErrFileNotFound {
		aelog.Warnf("Error while delete from index. %s: %v", name, err)
	}
	return
}

func (i *Index) add(file *File) (err error) {
	parts := i.explode(file.Name)
	if err = i.Root.Add(parts, file, 0); err != nil {
		return
	}
	atomic.AddInt64(&i.v, 1)
	atomic.AddInt64(&i.c, 1)
	return
}

func (i *Index) explode(name string) (parts []string) {
	return strings.Split(name, "/")
}
