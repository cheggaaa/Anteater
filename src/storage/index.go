package storage

import (
	"sync"
)

type Index struct {
	Files map[string]*File
	m *sync.Mutex
}


func (i *Index) Add(name string, file *File) {
	i.m.Lock()
	defer i.m.Unlock()
	i.Files[name] = file
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
	}
	return
}