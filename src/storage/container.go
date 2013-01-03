/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package storage

import (
	"os"
	"sync"
	//"sync/atomic"
	"syscall"
	"fmt"
	"time"
	"aelog"
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
	go func() {
		for {
			c.m.Lock()
			if len(c.Spaces) >= 200 {
				c.Clean()
			}
			c.m.Unlock()
			time.Sleep(time.Second * 5)
		}
	}()
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
		c.Count++
		c.ch = true
	}
	return
}

func (c *Container) getSpace(size int64, target int) (start int64, ok bool) {
	/*defer func(){
		if ok {
			c.Spaces.Sort()
			c.Spaces, c.MaxSpaceSize, c.Offset = c.Spaces.Join(c.Offset)
		}
	}()*/

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
				c.Offset += size
				start = c.Offset - size
				ok = true
			}
	}
	return
}

func (c *Container) Delete(f *File) {
	c.m.Lock()
	defer c.m.Unlock()
	c.Count--
	// is last file in container
	if f.Start + f.Size == c.Offset {
		c.Offset -= f.Size
	} else {
		s := &Space{f.Start, f.Size} 
		c.Spaces = append(c.Spaces, s)
	}
	if len(c.Spaces) < 200 {
		c.Clean()
	}
	c.ch = true	
}


/**
 * Maximum space available for new file
 */
func (c *Container) MaxSpace(target int) int64 {
	if target < TARGET_NEW {
		return c.MaxSpaceSize
	}
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

func (c *Container) Clean() {
	st := time.Now()
	aelog.Debugf("%d c. Start celan", c.Id)
	c.Spaces.Sort()
	c.Spaces, c.MaxSpaceSize, c.Offset = c.Spaces.Join(c.Offset)
	aelog.Debugf("%d c. Clean end for a %v", c.Id, time.Now().Sub(st))
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


