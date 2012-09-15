package storage

import (
	"time"
	"io"
	"fmt"
	"errors"
	"sync/atomic"
)

type File struct {
	CId int32
	Start int64
	Size int64
	Time time.Time
	Md5 []byte
	s *Storage
	c *Container
	openCount int32
	isDeleted bool
	willBeDeleted bool
}

func (f *File) Init(s *Storage) {
	f.s = s
	if f.c == nil {
		f.c = f.s.Containers.Get(f.CId)
	}
	if f.c == nil {
		panic(fmt.Sprintf("Can't init file from container %d", f.CId))
	}
}

func (f *File) Open() error {
	if f.isDeleted {
		return errors.New("File already deleted")
	}
	if f.willBeDeleted {
		return errors.New("File will be deleted")
	}
	atomic.AddInt32(&f.openCount, 1)
	return nil
}

func (f *File) GetReader() *io.SectionReader {
	return io.NewSectionReader(f.c.F, f.Start, f.Size)
}


func (f *File) WriteAt(b []byte, off int64) (int, error) {
	if off + int64(len(b)) > f.Size {
		panic("Can't write. Overflow allocated size")		
	}
	off = off + f.Start
	return f.c.F.WriteAt(b, off)
}

func (f *File) Delete() {
	if f.openCount == 0 {
		f.isDeleted = true
		f.c.Delete(f)
	} else {
		f.willBeDeleted = true
	}
}

func (f *File) Close() {
	v := atomic.AddInt32(&f.openCount, -1)
	if f.willBeDeleted && v == 0 {
		f.Delete()
	}
}
