package storage

import (
	"time"
	"io"
	"fmt"
)

type File struct {
	CId int32
	Start int64
	Size int64
	Time time.Time
	Md5 []byte
	s *Storage
	c *Container
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

func (f *File) GetReader() *io.SectionReader {
	return io.NewSectionReader(f.c.F, f.Start,f.Size)
}


func (f *File) WriteAt(b []byte, off int64) (int, error) {
	if off + int64(len(b)) > f.Size {
		panic("Can't write. Overflow allocated size")		
	}
	off = off + f.Start
	return f.c.F.WriteAt(b, off)
}

func (f *File) Delete() {
	f.c.Delete(f)
}
