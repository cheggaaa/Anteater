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
	"io"
	"config"
	"sync"
	"errors"
	"os"
	"time"
	"cnst"
	"dump"
	"fmt"
	"stats"
	"aelog"
	"utils"
)

const (
	TARGET_SPACE_EQ   = 0
	TARGET_SPACE_FREE = 1
	TARGET_NEW        = 2
)

type StorageDump struct {
	Index IndexDump
	Containers []ContainerDump
	Version string
	Time time.Time
}

type Storage struct {
	Index *Index
	Containers *Containers
	Conf *config.Config
	Stats *stats.Stats
	wm   *sync.Mutex
	sFuncLinks map[int]func()
}


func GetStorage(c *config.Config) (s *Storage) {
	s = &Storage{
		Conf : c,
		wm   : &sync.Mutex{},
	}
	aelog.Infof("Try restore from %s... ", s.DumpFilename())
	err, exists := s.Restore()
	if err != nil {
		if  ! exists {
			aelog.Infof("index does not exists, create new storage.. ")
			s.Create()
		} else {
			panic(err)
		}
	}
	s.Init()
	return
}


func (s *Storage) Init() {
	s.Stats = stats.New()
	if ! s.lock() {
		panic(errors.New("Anteater already running, or was crashed(( Check process or remove lock file"))
	}
	if s.Conf.DumpTime > 0 {
		go func() { 
				ch := time.Tick(s.Conf.DumpTime)
				for _ = range ch {
					func () {
						err := s.Dump()
						if err != nil {
							panic(err)
						}
					}()
				}
			}()
	}
	s.sFuncLinks = map[int]func(){
		TARGET_SPACE_EQ   : func() { s.Stats.Allocate.Replace.Add() },
		TARGET_SPACE_FREE : func() { s.Stats.Allocate.In.Add() },
		TARGET_NEW        : func() { s.Stats.Allocate.Append.Add() },
	}
	
	go func() {
		for {
			time.Sleep(time.Minute)
			s.Check()
		}
	}()
	
	return
}

func (s *Storage) Create() {
	s.Index = &Index{make(map[string]*File), &sync.Mutex{}, 0}
	s.Containers = &Containers{
		s : s,
		m : &sync.Mutex{},
		Containers : make(map[int32]*Container),
	}
	_, err := s.Containers.Create()
	if err != nil {
		panic(err)
	}
	return
}

func (s *Storage) Restore() (err error, exists bool) {
	data := new(StorageDump)
	err, exists = dump.LoadData(s.DumpFilename(), data)
	if err != nil {
		return
	}
	// restore containers
	containers := &Containers{
		s : s,
		m : &sync.Mutex{},	
	}
	containers.Containers = make(map[int32]*Container, len(data.Containers))
	for _, ct := range data.Containers {
		containers.Containers[ct.Id], err = ct.Restore(s)
		if err != nil {
			return
		}
		if ct.Id > containers.LastId {
			containers.LastId = ct.Id
		}
	}
	s.Containers = containers
	// restore index
	s.Index = data.Index.Restore(s)
	return 
}

func (s *Storage) Add(name string, r io.Reader, size int64) (f *File) {
	f = &File{
		Size: size,
		s : s,
		name : name,
		Time : time.Now(),
	}
	var ok bool
	defer func() {
		if ! ok {
			// if not added - remove file
			if f.CId != 0 {
				f.Delete()
			}
		}
	}()
	
	err := s.allocateFile(f)
	if err != nil {
		panic(err)
	}
	
	written, err := f.ReadFrom(r)
	
	if err != nil {
		panic(err)
	}
	
	if written != size {
		panic(fmt.Sprintf("Error while adding file. Requested size %d, but written only %d", size, written))
	}
	
	err = s.Index.Add(name, f)
	if err != nil {
		panic(err)
	}
	s.Stats.Counters.Add.Add()
	ok = true
	return
}

func (s *Storage) allocateFile(f *File) (err error) {
	var targets = []int{TARGET_SPACE_EQ, TARGET_SPACE_FREE, TARGET_NEW}
	var ok bool
	for _, target := range targets {
		for _, c := range s.Containers.Containers {
			if c.MaxSpace(target) >= f.Size {
				ok = c.Add(f, target)
				if ok {
					s.sFuncLinks[target]()
					return
				}
			}
		}
	}
	c, err := s.Containers.Create()
	if err != nil {
		return
	}
	ok = c.Add(f, TARGET_NEW)
	if ! ok {
		return errors.New("Can't allocate space")
	}
	s.Stats.Allocate.Append.Add()
	return
}

func (s *Storage) Get(name string) (f *File, ok bool) {
	f, ok = s.Index.Get(name)
	return
}

func (s *Storage) Delete(name string) (ok bool) {
	f, ok := s.Index.Delete(name)
	if ok {
		f.Delete()
	}
	return
}

func (s *Storage) Dump() (err error) {
	s.wm.Lock()
	s.Index.m.Lock()
	st := time.Now()
	containers := make([]ContainerDump, 0)
	
	for _, c := range s.Containers.Containers {
		containers = append(containers, c.DumpData())
	}
	dumpData := &StorageDump{
		Version : cnst.VERSION,
		Time : time.Now(),
		Containers : containers,
		Index : s.Index.DumpData(),
	}
	s.Index.m.Unlock()
	s.wm.Unlock()
	prep := time.Since(st)
	fname := s.DumpFilename()
	n, err := dump.DumpTo(fname, dumpData)
	if err != nil {
		return
	}
	tot := time.Since(st)
	s.Stats.Storage.DumpSize = int64(n)
	s.Stats.Storage.DumpTime = st
	s.Stats.Storage.DumpSaveTime = tot
	s.Stats.Storage.DumpLockTime = prep
	aelog.Debugf("Dump: %s bytes writed to %s for %v prep(%v)", utils.HumanBytes(int64(n)), fname, tot, prep)
	return
}

func (s *Storage) Drop() (err error) {
	s.Close()
	if s.Containers != nil {
		for _, c := range s.Containers.Containers {
			err = os.Remove(c.Filename())
			if err != nil {
				return err
			}
		}
	}
	os.Remove(s.DumpFilename())
	os.Remove(s.DumpFilename() + ".td")
	return
}

func (s *Storage) Close() {	
	s.unLock()
	if s.Containers != nil && s.Containers.Containers != nil {
		for _, c := range s.Containers.Containers {
			c.Close()
		}
	}	
}

func (s *Storage) DumpFilename() string {
	return s.Conf.DataPath + "index"
}

func (s *Storage) Check() {
	needNew := true
	for _, c := range s.Containers.Containers {
		if c.MaxSpace(TARGET_NEW) > s.Conf.MinEmptySpace {
			needNew = false
			break
		}
	}
	if needNew {
		_, err := s.Containers.Create()
		if err != nil {
			aelog.Warnln("Error while create container: ", err)
		}
	}
}

func (s *Storage) GetStats() *stats.Stats {	
	s.Stats.Refresh()
	
	s.Stats.Storage.ContainersCount = len(s.Containers.Containers)
	s.Stats.Storage.FilesCount = len(s.Index.Files)
	s.Stats.Storage.IndexVersion = s.Index.Version()
	s.Stats.Storage.FilesSize = 0
	s.Stats.Storage.TotalSize = 0
	s.Stats.Storage.HoleCount = 0
	s.Stats.Storage.HoleSize = 0
	for _, c := range s.Containers.Containers {
		s.Stats.Storage.TotalSize += c.Size
		hc, hs := c.Spaces.Stats()
		s.Stats.Storage.HoleCount += hc
		s.Stats.Storage.HoleSize += hs
	}
	
	for n, _ := range s.Index.Files {
		f, ok := s.Index.Get(n)
		if ok {
			s.Stats.Storage.FilesSize += f.Size
		}
	}
		
	return s.Stats
}


func (s *Storage) lock() bool {
	f, err := os.Open(s.Conf.DataPath + "lock")
	if err == nil {
		f.Close()
		return false
	}
	f, err = os.Create(s.Conf.DataPath + "lock")
	if err != nil {
		fmt.Println(err)
		return false
	}
	f.WriteString(cnst.VERSION)
	f.Close()
	return err == nil
}

func (s *Storage) unLock() {
	os.Remove(s.Conf.DataPath + "lock")
}


func (s *Storage) CheckMD5() (result map[string]bool) {
	result = make(map[string]bool)
	
	for n, _ := range s.Index.Files {
		f, ok := s.Get(n)
		if ok {
			aelog.Debugf("Checking md5 for %s\n", n)
			result[n] = f.CheckMd5()
			if ! result[n] {
				aelog.Infof("File %s has mismatched md5\n", n)
			}
		}
	}
	return
}
