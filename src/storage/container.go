package storage

import (
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"fmt"
)

type Container struct {
	Id	   int32
	F     *os.File
	Size   int64
	Offset int64
	Count  int64
	LastId int64
	Spaces Spaces
	MaxSpaceSize int64
	ch    bool
	m     *sync.Mutex
	s	  *Storage
}

type ContainerDumpData struct {
	Id     int32
	Path   string
	Size   int64
	Offset int64
	Count  int64
	LastId int64
	Spaces Spaces
	MaxSpaceSize int64
}


func (c *Container) Init() (err error) {
	c.m = &sync.Mutex{}
	err = c.Open()
	return
}

func (c *Container) Create() (err error) {
	err = c.Init()
	if err != nil {
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	c.Size = c.s.Conf.ContainerSize
	err = c.allocate()
	return
}

func (c *Container) Open() (err error) {
	c.F, err = os.OpenFile(c.Filename(), os.O_RDWR | os.O_CREATE, 0666)
	return
}

func (c *Container) Close() {
	c.F.Close()
	return
}

func (c *Container) Filename() string {
	return fmt.Sprintf("%s/data.%d", c.s.Conf.DataPath, c.Id)
}

func (c *Container) Add(size int64, target int) (f *File, ok bool) {
	return
}

func (c *Container) Delete(f *File) {
	atomic.AddInt64(&c.Count, -1)
	c.m.Lock()
	defer c.m.Unlock()
	if f.Start + f.Size == c.Offset {
		atomic.AddInt64(&c.Offset, -f.Size)
	} else {
		c.Spaces = append(c.Spaces, &Space{f.Start, f.Size})
		c.Spaces.Sort()
	}
	c.ch = true	
}

func (c *Container) allocate() (err error) {
	err = syscall.Fallocate(int(c.F.Fd()), 0, 0, c.Size)
	return
}


