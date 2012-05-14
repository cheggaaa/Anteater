package anteater

import (
	"sync"
	"time"
	//"crypto/md5"
	"fmt"
)

type FileInfo struct {
	Id          int64
	ContainerId int
	Start       int64
	Size        int64
	C           int64
	T			int64
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
	info := &FileInfo{f.Id, f.C.Id, f.Start, f.Size, 0, time.Now().Unix()}
	Index[name] = info
	return info
}


func IndexGet(name string) (*FileInfo, bool) {
	if Index[name] == nil {
		return nil, false
	} 
	return Index[name], true
}

func (f *FileInfo) ETag () string {
	//h := md5.New()
	//fmt.Fprintf(h, "%d:%d:%d", f.Id, f.ContainerId, f.T)
	return fmt.Sprintf("%x%x%x", f.ContainerId, f.T, f.Id)
}
