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
	Spaces Spaces
	MaxSpaceSize int64
	ch    bool
	m     *sync.Mutex
	s	  *Storage
}

type ContainerDump struct {
	Id     int32
	Size   int64
	Offset int64
	Count  int64
	Spaces Spaces
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
	return fmt.Sprintf("%sdata.%d", c.s.Conf.DataPath, c.Id)
}

func (c *Container) Add(f *File, target int) (ok bool) {
	c.m.Lock()
	defer c.m.Unlock()
	size := f.Size
	if size > c.Size {
		return
	}
	start, ok := c.getSpace(size, target)
	if ok {
		f.Start = start
		f.CId = c.Id
		f.c = c
		atomic.AddInt64(&c.Count, 1)
		c.ch = true
	}
	return
}

func (c *Container) getSpace(size int64, target int) (start int64, ok bool) {
	defer func(){
		if ok {
			c.Spaces.Sort()
			c.Spaces, c.MaxSpaceSize, c.Offset = c.Spaces.Join(c.Offset)
		}
	}()

	switch (target) {
		case TARGET_SPACE_EQ, TARGET_SPACE_FREE:
			if c.MaxSpaceSize >= size {
				start, ok = c.Spaces.Get(size, target)
				return
			} else {
				return
			}
		case TARGET_NEW:
			if c.Offset + size <= c.Size {
				o := atomic.AddInt64(&c.Offset, size)
				start = o - size
				ok = true
			}
	}
	return
}

func (c *Container) Delete(f *File) {
	atomic.AddInt64(&c.Count, -1)
	c.m.Lock()
	defer c.m.Unlock()
	o := c.Offset
	if f.Start + f.Size == c.Offset {
		o = atomic.AddInt64(&c.Offset, -f.Size)
	} else {
		s := &Space{f.Start, f.Size} 
		c.Spaces = append(c.Spaces, s)
	}
	c.Spaces.Sort()
	c.Spaces, c.MaxSpaceSize, c.Offset = c.Spaces.Join(o)
	c.ch = true	
}


/**
 * Maximum space available for new file
 */
func (c *Container) MaxSpace() int64 {
	var spaceSize int64 = c.Size - c.Offset
	if c.MaxSpaceSize > spaceSize {
		return c.MaxSpaceSize
	}
	return spaceSize
}

func (c *Container) allocate() (err error) {
	err = syscall.Fallocate(int(c.F.Fd()), 0, 0, c.Size)
	return
}

/**
 * Return dump data
 */
func (c *Container) DumpData() (dump ContainerDump) {
	c.m.Lock()
	defer c.m.Unlock()
	dump.Count = c.Count
	dump.Id = c.Id
	dump.Size = c.Size
	dump.Offset = c.Offset
	dump.Spaces = c.Spaces
	return
}

func (cd *ContainerDump) Restore(s *Storage) (*Container, error) {
	c := &Container {
		Id : cd.Id,
		Count : cd.Count,
		Size : cd.Size,
		Offset : cd.Offset,
		Spaces : cd.Spaces,
		s : s,
	}
	err := c.Init()
	if err != nil {
		return nil, err
	}
	return c, nil
}


