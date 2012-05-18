package anteater

import (
	"errors"
	"io"
)


var Targets = []int{TARGET_SPACE_EQ, TARGET_SPACE_FREE, TARGET_NEW}

func GetFile(name string, size int64) (*File, *FileInfo, error) {
	GetFileLock.Lock()
	defer GetFileLock.Unlock()
	for _, target := range(Targets) {
		for _, c := range(FileContainers) {
			if c.MaxSpace() >= size {
				f, err := c.Allocate(size, target)
				if err == nil {
					fi := IndexSet(name, f)
					AllocCn.CTarget(target)
					return f, fi, nil
				}
			}
		}
	}
	cId, err := NewContainer(DataPath)
	if err != nil {
		return nil, nil, err
	}
	 
	f, err := FileContainers[cId].Allocate(size, TARGET_NEW)
	if err != nil {
		return nil, nil, err
	}
	
	fi := IndexSet(name, f)
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

