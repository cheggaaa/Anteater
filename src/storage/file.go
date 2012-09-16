package storage

import (
	"time"
	"io"
	"fmt"
	"errors"
	"sync/atomic"
	"strconv"
	"mime"
	"path/filepath"
	"net/http"
)

type File struct {
	CId int32
	Start int64
	Size int64
	Time time.Time
	Md5 []byte
	s *Storage
	c *Container
	name string
	openCount int32
	isDeleted bool
	willBeDeleted bool
	ctype string
}

type FileDump struct {
	CId int32
	Start int64
	Size int64
	Time time.Time
	Md5 []byte
}

func (f *File) Init(s *Storage, name string) {
	f.s = s
	if f.c == nil {
		f.c = f.s.Containers.Get(f.CId)
	}
	if f.c == nil {
		panic(fmt.Sprintf("Can't init file from container %d", f.CId))
	}
	f.name = name
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

func (f *File) ETag() string {
	return strconv.FormatInt(int64(f.CId), 36) +"."+strconv.FormatInt(f.Time.Unix(), 36)+"." +strconv.FormatInt(f.Start, 36)  +  "."+strconv.FormatInt(f.Size, 36);
}

// Return content type file or application/octed-stream if can't
func (f *File) ContentType() (ctype string) {
	ctype = f.ctype
	if ctype == "" {
		ctype = mime.TypeByExtension(filepath.Ext(f.name))
		if ctype == "" {
			var buf [1024]byte
			n, _ := io.ReadFull(f.GetReader(), buf[:])
			b := buf[:n]
			ctype = http.DetectContentType(b)
		}
		f.ctype = ctype
	}	
	return
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

func (f *File) DumpData() (dump FileDump) {
	dump.CId = f.CId
	dump.Md5 = f.Md5
	dump.Size = f.Size
	dump.Start = f.Start
	dump.Time = f.Time
	return
}
