package fire

import (
	"sync"
)

type FileInfo struct {
	Id          int64
	ContainerId int
	Start       int64
	Size        int64
	C           int64
}

var Index map[string]*FileInfo
var IndexMutex *sync.Mutex = &sync.Mutex{}

func init() {
	IndexMutex.Lock()
	defer IndexMutex.Unlock()
	Index = make(map[string]*FileInfo, 1000)
}

func IndexAdd(name string, f *File) *FileInfo {
	IndexMutex.Lock()
	defer IndexMutex.Unlock()
	info := &FileInfo{f.Id, f.C.Id, f.Start, f.Size, 0}
	Index[name] = info
	return info
}


func IndexGet(name string) (*FileInfo, bool) {
	if Index[name] == nil {
		return nil, false
	} 
	return Index[name], true
}
