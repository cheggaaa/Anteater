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
	"testing"
)

var names = []string{
	"foo",
	"foo/bar",
	"foo/bar/baz",
	"foo/bar/bar",
	"foo/bar/foo",
	"foo/zzz",
	"lo/olo/lol.lol",
	"a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/r/s/t/u/v/w/x/y/z",
}

var I = &Index{}

func TestAdd(t *testing.T) {
	I.Init()
	for _, name := range names {
		if err := I.Add(&File{Name:name}); err != nil {
			t.Errorf("Can't add file. %s: %v", name, err)
		}
	}
	if I.Count() != int64(len(names)) {
		t.Errorf("Count mismatched: %d vs %d", I.Count(), len(names))
	}
}

func TestGet(t *testing.T) {
	for _, name := range names {
		f, ok := I.Get(name)
		if ! ok {
			t.Errorf("Can't get file: %s", name)
		}
		if f.Name != name {
			t.Errorf("Files names not equals: %s vs %s", f.Name, name)
		}
	}
}

func TestList(t *testing.T) {
	files, err := I.List("", 0)
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != len(names) {
		t.Errorf("Count list mismatched: %d vs %d", len(files), len(names))
	}
	
	files, err = I.List("foo", 0)
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != 5 {
		t.Errorf("Count list mismatched: %d vs %d\n%v", len(files), 5, files)
	}
	
	files, err = I.List("foo/bar", 0)
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Count list mismatched: %d vs %d\n%v", len(files), 3, files)
	}
	
	files, err = I.List("a", 0)
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Count list mismatched: %d vs %d\n%v", len(files), 1, files)
	}
	
	files, err = I.List("z", 0)
	if err != ErrFileNotFound {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Count list mismatched: %d vs %d\n%v", len(files), 0, files)
	}
	I.Root.Print("")
}

func TestDelete(t *testing.T) {
	i := int64(len(names))
	for _, name := range names {
		i--
		f, ok := I.Delete(name)
		if ! ok {
			t.Errorf("Can't delete file: %s", name)
		}
		if f.Name != name {
			t.Errorf("Files names not equals: %s vs %s", f.Name, name)
		}
		if I.Count() != i {
			t.Errorf("Count mismatched: %d vs %d", I.Count(), i)
		}
	}
	I.Root.Print("")
}

func TestRename(t *testing.T) {
	for i, name := range names {
		if i % 2 == 0 {
			if err := I.Add(&File{Name:name}); err != nil {
				t.Errorf("Can't add file. %s: %v", name, err)
			}
		} else {
			_, err := I.Rename(names[i-1], name)
			if err != nil {
				t.Errorf("Can't rename file. %s: %v", name, err)
			}
		}
	}
	for i, name := range names {
		_, ok := I.Get(name)
		if i % 2 == 0 {
			if ok {
				t.Errorf("File exists")
			}
		} else {
			if ! ok {
				t.Errorf("File not exists")
			}
		}
	}
	I.Root.Print("")
}

