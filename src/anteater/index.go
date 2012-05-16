package anteater

import (
	"time"
	"fmt"
	"io"
)

type FileInfo struct {
	Id          int64
	ContainerId int32
	Start       int64
	Size        int64
	C           int64
	T			int64
}



func init() {
	IndexLock.Lock()
	defer IndexLock.Unlock()
	Index = make(map[string]*FileInfo, 1000)
}

func IndexSet(name string, f *File) *FileInfo {
	IndexLock.Lock()
	defer IndexLock.Unlock()
	info := &FileInfo{f.Id, f.C.Id, f.Start, f.Size, 0, time.Now().Unix()}
	Index[name] = info
	return info
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
