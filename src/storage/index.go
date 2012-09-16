package storage

import (
	"sync"
	"sync/atomic"
)

type Index struct {
	Files map[string]*File
	m *sync.Mutex
	v int64
}

type IndexDump struct {
	F map[string]FileDump
	V int64
}

func (i *Index) Add(name string, file *File) {
	i.m.Lock()
	defer i.m.Unlock()
	i.Files[name] = file
	atomic.AddInt64(&i.v, 1)
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
		atomic.AddInt64(&i.v, 1)
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