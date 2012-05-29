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

package anteater

import (
	"os"
	"syscall"
	"sync"
	"sync/atomic"
	"errors"
	"strconv"
)

const (
	TARGET_SPACE_EQ   = 0
	TARGET_SPACE_FREE = 1
	TARGET_NEW        = 2
)

type Container struct {
	Id	   int32
	F     *os.File
	Size   int64
	Offset int64
	Count  int64
	LastId int64
	WLock *sync.Mutex
	Spaces Spaces
	MaxSpaceSize int64
	Ch    bool
}

type ContainerDumpData struct {
	Id     int32
	Size   int64
	Offset int64
	Count  int64
	LastId int64
	Spaces Spaces
	MaxSpaceSize int64
}

/**
 *	Create new container and return id
 */
func NewContainer(path string) (int32, error) {
	id := ContainerNextId()
	size := Conf.ContainerSize
	path += "." + strconv.FormatInt(int64(id), 10)
	
	Log.Infoln("Create new container", path, "...")
	
	f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		return 0, err
	}	
	c := &Container{ContainerLastId, f, size, 0, 0, 0, &sync.Mutex{}, make([]*Space, 0), 0, true}

	if err != nil {
		return 0, err
	}
	
	Log.Debugln("Try allocate", size, "bytes");
		
	err = c.Init()
	if err != nil {
		return 0, err
	}	
	Log.Debugln("Container", c.Id, "created");	
	FileContainers[c.Id] = c
	return c.Id, err
}

/**
 * Create container from ContainerDumpData
 */
func ContainerFromData(data *ContainerDumpData) (*Container, error) {
	Log.Infoln("Init container", data.Id);	
	path := DataPath + "." + strconv.FormatInt(int64(data.Id), 10)
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	c := &Container{data.Id, f, data.Size, data.Offset, data.Count, data.LastId, &sync.Mutex{}, data.Spaces, data.MaxSpaceSize, false}
	return c, nil
}

/**
 * Generate new container id
 */
func ContainerNextId() int32 {
	return atomic.AddInt32(&ContainerLastId, 1)
}

/**
 * Init container, must call after create
 * Allocate file for container
 */
func (c *Container) Init() error {
	err := syscall.Fallocate(int(c.F.Fd()), 0, 0, c.Size)
	if err != nil {
		return err
	}
	return nil
}

func (c *Container) Path() string {
	return DataPath + "." + strconv.FormatInt(int64(c.Id), 10)
}

/**
 * Try Allocate space for new file
 */
func (c *Container) Allocate(size int64, target int) (*File, error) {
	c.WLock.Lock()
	defer c.WLock.Unlock()
	if size > c.Size {
		return nil, errors.New("Can't allocate space in container " + c.Path())
	}
	start, err := c.GetSpace(size, target)
	if err != nil {
		return nil, err
	}
	id := atomic.AddInt64(&c.LastId, 1)
	atomic.AddInt64(&c.Count, 1)
	c.Ch = true
	return &File{id, c, start, size}, nil
}

/**
 * Delete file
 */
func (c *Container) Delete(info *FileInfo) {
	atomic.AddInt64(&c.Count, -1)
	c.WLock.Lock()
	defer c.WLock.Unlock()
	if info.Id == c.LastId {
		atomic.AddInt64(&c.Offset, 0 - info.Size)
	} else {
		c.Spaces = append(c.Spaces, &Space{info.Start, info.Size})
		c.Spaces.Sort()
	}
	c.Ch = true	
}

/**
 * Allocate sace for new files
 */
func (c *Container) GetSpace(size int64, target int) (int64, error) {
	switch (target) {
		case TARGET_SPACE_EQ, TARGET_SPACE_FREE:
			if c.MaxSpaceSize >= size {
				return c.Spaces.Get(size, target)
			} else {
				return 0, errors.New("Can't allocate space")
			}
		case TARGET_NEW:
			if c.Offset + size <= c.Size {
				o := atomic.AddInt64(&c.Offset, size)
				return o - size, nil
			} else {
				return 0, errors.New("Can't allocate space")
			}
	}
	return 0, errors.New("Undefined target")
}

/**
 * Clean container
 */
func (c *Container) Clean() {	
	if c.HasChanges() {
		Log.Debugln("Start clean container", c.Id);
		c.WLock.Lock()
		c.Spaces, c.MaxSpaceSize = c.Spaces.Join()
		c.Ch = false
		c.WLock.Unlock()
	}	
}

/**
 * Return true if container has changes after last dump
 */
func (c *Container) HasChanges() bool {
	return c.Ch
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

/**
 * Return data for dump
 */
func (c *Container) GetDumpData() *ContainerDumpData {
	c.WLock.Lock()
	defer c.WLock.Unlock()
	return &ContainerDumpData{c.Id, c.Size, c.Offset, c.Count, c.LastId, c.Spaces, c.MaxSpaceSize}
}
