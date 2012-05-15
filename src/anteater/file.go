package anteater

import (
	"errors"
	"sync"
	"io"
)

const (
	TARGET_SPACE = 0
	TARGET_NEW   = 1
)

var Targets = []int{TARGET_SPACE, TARGET_NEW}

var GetFileLock *sync.Mutex = &sync.Mutex{}


func GetFile(name string, size int64) (*File, *FileInfo, error) {
	GetFileLock.Lock()
	defer GetFileLock.Unlock()
	for _, target := range(Targets) {
		for _, c := range(FileContainers) {
			if c.MaxSpace() >= size {
				f, err := c.New(size, target)
				if err == nil {
					fi := IndexAdd(name, f)
					return f, fi, nil
				}
			}
		}
	}
	c, err := NewContainer(DataPath, CSize)	
	var f *File
	if err != nil {
		return nil, nil, err
	} else {
		FileContainers[c.Id] = c
		f, err = c.New(size, TARGET_NEW)
		if err != nil {
			return nil, nil, err
		}
	}
	fi := IndexAdd(name, f)
	return f, fi, nil
}


type File struct {
	Id    int64
	C    *Container
	Start int64
	Size  int64
}

func (f *File) WriteAt(b []byte, off int64) (int, error) {
	if off + int64(len(b)) > f.Size {
		return 0, errors.New("Can't write. Overflow allocated size")		
	}
	off = off + f.Start
	return f.C.F.WriteAt(b, off)
}

func (f *File) GetReader() *io.SectionReader {
	return io.NewSectionReader(f.C.F, f.Start,f.Size)
}

