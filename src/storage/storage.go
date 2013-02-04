package storage

import (
	"config"
	"stats"
	"os"
	"errors"
	"path/filepath"
	"dump"
	"sync"
	"aelog"
	"fmt"
	"io"
	"time"
	"encoding/gob"
)

func init() {
	gob.Register(&File{})
	gob.Register(&Hole{})
}

type Storage struct {
	Conf *config.Config
	Index *Index
	LastContainerId int64
	Stats *stats.Stats
	Containers map[int64]*Container
	
	m *sync.Mutex
}

func (s *Storage) Init(c *config.Config) {
	s.Conf = c
	s.m = new(sync.Mutex)
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
	if ! info.IsDir() {
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
		_, err = s.createContainer()
	}
	return
}


func (s *Storage) Add(name string, r io.Reader, size int64) (f *File, err error) {
	f = &File{
		Name : name,
		FSize : size,
		Time : time.Now(),
	}
	var target int
	defer func() {
		if err != nil {
			f.Delete()
		} else {
			s.Stats.Allocate.Replace.Add()
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
	for _, target = range Targets {	
		for _, c := range s.Containers {
			if ok := c.Allocate(f, target); ok {
				return
			}
		}
	}	
	// create new container
	c, err := s.createContainer()
	if err != nil {
		return
	}
	if ok := c.Allocate(f, ALLOC_APPEND); ! ok {
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
	s.Stats.Storage.FilesCount = int64(len(s.Index.Files))
	s.Stats.Storage.IndexVersion = s.Index.Version()
	s.Stats.Storage.FilesSize = 0
	s.Stats.Storage.TotalSize = 0
	s.Stats.Storage.HoleCount = 0
	s.Stats.Storage.HoleSize = 0
	for _, c := range s.Containers {
		s.Stats.Storage.TotalSize += c.Size
		hc, hs := c.holeIndex.Count, c.holeIndex.Size
		s.Stats.Storage.HoleCount += hc
		s.Stats.Storage.HoleSize += hs
		s.Stats.Storage.FilesSize += c.FileSize
	}	
	return s.Stats
}

func (s *Storage) CheckMD5() map[string]bool {
	return nil
}

func (s *Storage) Close() {

}

func (s *Storage) restoreContainer(path string) (err error) {
	aelog.Debugf("Restore container from %s..", path)
	container := &Container{}
	if err, _ = dump.LoadData(path, container); err != nil {
		return err
	}
	if container.Created {
		if err = container.Init(s); err != nil {
			return
		}
		s.m.Lock()
		if container.Id > s.LastContainerId {
			s.LastContainerId = container.Id
		}
		s.Containers[container.Id] = container
		s.m.Unlock()
		aelog.Debugf("Container %d restored. %d files found", container.Id, container.FileCount)
	} else {
		return fmt.Errorf("Can't restore container from %s", path)
	}
	return
}

func (s *Storage) createContainer() (c *Container, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	s.LastContainerId++
	s.Containers[s.LastContainerId] = &Container{
		Id : s.LastContainerId,
	}
	return s.Containers[s.LastContainerId], s.Containers[s.LastContainerId].Init(s)
}