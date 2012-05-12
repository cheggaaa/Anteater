package anteater

import (
	"os"
	"syscall"
	"sync"
	"sync/atomic"
	"errors"
	"strconv"
)

type Container struct {
	Id	   int
	F     *os.File
	Path   string
	Size   int64
	Offset int64
	Count  int64
	LastId int64
	M     *sync.Mutex
	Spaces Spaces
	MaxSpaceSize int64
	Ch    bool
}

type ContainerDumpData struct {
	Id     int
	Path   string
	Size   int64
	Offset int64
	Count  int64
	LastId int64
	Spaces Spaces
	MaxSpaceSize int64
}


var FileContainers map[int]*Container = make(map[int]*Container)
var ContainerLastId int
var CreateMutex *sync.Mutex = &sync.Mutex{}

func NewContainer(path string, size int64) (*Container, error) {
	CreateMutex.Lock()
	defer CreateMutex.Unlock()
	ContainerLastId++
	path += "." + strconv.FormatInt(int64(ContainerLastId), 10)
	
	Log.Infoln("Create new container", path, "...")
	
	f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}	
	c := &Container{ContainerLastId, f, path, 0, 0, 0, 0, &sync.Mutex{}, make([]*Space, 0), 0, true}

	if err != nil {
		return nil, err
	}
	
	Log.Debugln("Try allocate", size, "bytes");
		
	err = c.Allocate(size)
	if err != nil {
		return nil, err
	}
	
	Log.Debugln("Container", c.Id, "created");	
	return c, err
}

func ContainerFromData(data *ContainerDumpData) (*Container, error) {
	Log.Infoln("Init container", data.Id);	
	f, err := os.OpenFile(data.Path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	c := &Container{data.Id, f, data.Path, data.Size, data.Offset, data.Count, data.LastId, &sync.Mutex{}, data.Spaces, data.MaxSpaceSize, false}
	return c, nil
}

func (c *Container) Allocate(size int64) error {
	err := syscall.Fallocate(int(c.F.Fd()), 0, 0, size)
	if err != nil {
		return err
	}
	c.Size = size
	return nil
}

func (c *Container) New(size int64, target int) (*File, error) {
	start, err := c.GetSpace(size, target)
	if err != nil {
		return nil, err
	}
	id := atomic.AddInt64(&c.LastId, 1)
	atomic.AddInt64(&c.Count, 1)
	if start + size > c.Size {
		return nil, errors.New("Can't allocate space in container " + c.Path)
	}
	c.Ch = true
	return c.Get(id, start, size), nil
}

func (c *Container) Get(id, start, size int64) (*File) {
	return &File{id, c, start, size, 0}
}

func (c *Container) Delete(id, start, size int64) {
	atomic.AddInt64(&c.Count, -1)
	c.M.Lock()
	if id == c.LastId {
		atomic.AddInt64(&c.Offset, 0 - size)
	} else {
		c.Spaces = append(c.Spaces, &Space{start, size})
		c.Spaces.Sort()
	}
	c.Ch = true
	c.M.Unlock()
}

func (c *Container) GetSpace(size int64, target int) (int64, error) {
	switch (target) {
		case TARGET_SPACE:
			if c.MaxSpaceSize >= size {
				return c.Spaces.Get(size)
			} else {
				return 0, errors.New("Can't allocate space")
			}
		case TARGET_NEW:
			o := atomic.AddInt64(&c.Offset, size)
			return o - size, nil
	}
	return 0, errors.New("Undefined target")
}

func (c *Container) Clean() {
	Log.Debugln("Start clean container", c.Id);
	if c.HasChanges() {
		c.M.Lock()
		c.Spaces, c.MaxSpaceSize = c.Spaces.Join()
		c.Ch = false
		c.M.Unlock()
	}	
}

func (c *Container) HasChanges() bool {
	return c.Ch
}

func (c *Container) MaxSpace() int64 {
	var spaceSize int64 = c.Size - c.Offset
	if c.MaxSpaceSize > spaceSize {
		return c.MaxSpaceSize
	}
	return spaceSize
}

func (c *Container) GetDumpData() ContainerDumpData {
	c.M.Lock()
	defer c.M.Unlock()
	return ContainerDumpData{c.Id, c.Path, c.Size, c.Offset, c.Count, c.LastId, c.Spaces, c.MaxSpaceSize}	
}
