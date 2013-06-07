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
	files, err := I.List("")
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != len(names) {
		t.Errorf("Count list mismatched: %d vs %d", len(files), len(names))
	}
	
	files, err = I.List("foo")
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != 5 {
		t.Errorf("Count list mismatched: %d vs %d\n%v", len(files), 5, files)
	}
	
	files, err = I.List("foo/bar")
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Count list mismatched: %d vs %d\n%v", len(files), 3, files)
	}
	
	files, err = I.List("a")
	if err != nil {
		t.Errorf("Can't get list: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Count list mismatched: %d vs %d\n%v", len(files), 1, files)
	}
	
	files, err = I.List("z")
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

