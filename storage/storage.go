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
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/cheggaaa/Anteater/aelog"
	"github.com/cheggaaa/Anteater/config"
	"github.com/cheggaaa/Anteater/dump"
	"github.com/cheggaaa/Anteater/stats"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func init() {
	gob.Register(&File{})
	gob.Register(&Hole{})
}

type Storage struct {
	Conf            *config.Config
	Index           *Index
	LastContainerId int64
	Stats           *stats.Stats
	Containers      map[int64]*Container

	m sync.RWMutex
}

func (s *Storage) Init(c *config.Config) {
	s.Conf = c
	s.Index = new(Index)
	s.Index.Init()
	s.Containers = make(map[int64]*Container)
	s.Stats = stats.New()
}

func (s *Storage) Open() (err error) {
	aelog.Debugf("Try open %s..", s.Conf.DataPath)
	dir, err := os.Open(s.Conf.DataPath)
	if err != nil {
		return
	}
	info, err := dir.Stat()
	if err != nil {
		return
	}
	if !info.IsDir() {
		return errors.New("Data path must be dir")
	}

	wg := &sync.WaitGroup{}
	files, err := dir.Readdir(-1)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".index" {
			wg.Add(1)
			go func(name string) {
				name = s.Conf.DataPath + name
				e := s.restoreContainer(name)
				if e != nil {
					err = e
				}
				wg.Done()
			}(file.Name())
		}
	}
	wg.Wait()

	if err != nil {
		return
	}

	if len(s.Containers) == 0 {
		aelog.Info("Create first container")
		if _, err = s.createContainer(); err != nil {
			return
		}
	}

	go func() {
		if s.Conf.DumpTime > 0 {
			for {
				time.Sleep(s.Conf.DumpTime)
				s.Dump()
			}
		}
	}()

	return
}

func (s *Storage) Add(name string, r io.Reader, size int64) (f *File, err error) {
	f = &File{
		Name:  name,
		FSize: size,
		Time:  time.Now(),
	}
	var target int
	defer func() {
		if err != nil {
			f.Delete()
		} else {
			switch target {
			case ALLOC_REPLACE:
				s.Stats.Allocate.Replace.Add()
			case ALLOC_APPEND:
				s.Stats.Allocate.Append.Add()
			case ALLOC_INSERT:
				s.Stats.Allocate.In.Add()
			}
			s.Stats.Counters.Add.Add()
		}
	}()

	if size <= 0 {
		err = fmt.Errorf("Can't allocate 0-size for %s", name)
		return
	}

	// allocate
	target, err = s.allocate(f)
	if err != nil {
		return
	}

	// write
	w, err := f.ReadFrom(r)
	if err != nil {
		return
	}

	// check
	if w != size {
		err = fmt.Errorf("Requested %d bytes, but writed only %d", size, w)
		return
	}

	// add to index
	err = s.Index.Add(f)
	return
}

func (s *Storage) allocate(f *File) (target int, err error) {
	s.m.RLock()
	for _, target = range Targets {
		for _, c := range s.Containers {
			if ok := c.Allocate(f, target); ok {
				s.m.RUnlock()
				return
			}
		}
	}
	s.m.RUnlock()
	// create new container
	c, err := s.createContainer()
	if err != nil {
		return
	}
	if ok := c.Allocate(f, ALLOC_APPEND); !ok {
		return 0, fmt.Errorf("Can't allocate space!")
	}
	target = ALLOC_APPEND
	return
}

func (s *Storage) Get(name string) (f *File, ok bool) {
	return s.Index.Get(name)
}

func (s *Storage) Delete(name string) (ok bool) {
	f, ok := s.Index.Delete(name)
	if ok {
		f.Delete()
	}
	return
}

func (s *Storage) DeleteChilds(name string) (ok bool) {
	names, err := s.Index.List(name, 0)
	if err != nil {
		aelog.Warnln("Can't get file list:", err)
		return
	}
	for _, name := range names {
		if s.Delete(name) && !ok {
			ok = true
		}
	}
	return
}

func (s *Storage) Dump() {
	s.m.Lock()
	defer s.m.Unlock()
	for _, c := range s.Containers {
		err := c.Dump()
		if err != nil {
			aelog.Warnf("Can't dump container %d: %v", c.Id, err)
		}
	}
}

func (s *Storage) GetStats() *stats.Stats {
	s.Stats.Refresh()

	s.Stats.Storage.ContainersCount = len(s.Containers)
	s.Stats.Storage.FilesCount = s.Index.Count()
	s.Stats.Storage.IndexVersion = s.Index.Version()
	s.Stats.Storage.FilesSize = 0
	s.Stats.Storage.TotalSize = 0
	s.Stats.Storage.HoleCount = 0
	s.Stats.Storage.HoleSize = 0
	s.Stats.Storage.FilesRealSize = 0
	for _, c := range s.Containers {
		c.m.Lock()
		s.Stats.Storage.TotalSize += c.Size
		hc, hs := c.holeIndex.Count, c.holeIndex.Size
		s.Stats.Storage.HoleCount += hc
		s.Stats.Storage.HoleSize += hs
		s.Stats.Storage.FilesSize += c.FileSize
		s.Stats.Storage.FilesRealSize += c.FileRealSize
		c.m.Unlock()
	}
	return s.Stats
}

func (s *Storage) Check() (err error) {
	s.m.RLock()
	defer s.m.RUnlock()
	for _, c := range s.Containers {
		if err = c.Check(); err != nil {
			return
		}
	}
	return
}

func (s *Storage) Close() {
	s.Dump()
	s.m.RLock()
	defer s.m.RUnlock()
	for _, c := range s.Containers {
		c.Close()
	}
}

func (s *Storage) Drop() {
	s.Close()
	for _, c := range s.Containers {
		os.Remove(c.fileName())
		os.Remove(c.indexName())
	}
}

func (s *Storage) restoreContainer(path string) (err error) {
	aelog.Debugf("Restore container from %s..", path)
	rr, err, _ := dump.LoadData(path)
	if err != nil {
		return err
	}
	tp, err := rr.B.ReadByte()
	if err != nil {
		return
	}
	if tp != 3 {
		return fmt.Errorf("unexpected container type: %v", tp)
	}

	id, err := binary.ReadUvarint(rr.B)
	if err != nil {
		return
	}
	sz, err := binary.ReadUvarint(rr.B)
	if err != nil {
		return
	}
	fc, err := binary.ReadUvarint(rr.B)
	if err != nil {
		return
	}
	fs, err := binary.ReadUvarint(rr.B)
	if err != nil {
		return
	}
	frs, err := binary.ReadUvarint(rr.B)
	if err != nil {
		return
	}
	cr, err := rr.B.ReadByte()
	if err != nil {
		return
	}
	container := &Container{
		Id:           int64(id),
		Size:         int64(sz),
		FileCount:    int64(fc),
		FileSize:     int64(fs),
		FileRealSize: int64(frs),
		Created:      cr == 11,
	}

	if container.Created {
		if err = container.Init(s, rr); err != nil {
			return
		}
		s.m.Lock()
		if container.Id > s.LastContainerId {
			s.LastContainerId = container.Id
		}
		s.Containers[container.Id] = container
		s.m.Unlock()
		aelog.Debugf("Container %d restored. %d (%d) files (holes) found", container.Id, container.FileCount, container.holeIndex.Count)
		//container.holeIndex.Print()
		//container.Print()
	} else {
		return fmt.Errorf("Can't restore container from %s", path)
	}
	return
}

func (s *Storage) createContainer() (c *Container, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	s.LastContainerId++
	c = &Container{
		Id: s.LastContainerId,
	}
	err = c.Init(s, nil)
	if err == nil {
		s.Containers[s.LastContainerId] = c
	}
	return
}
