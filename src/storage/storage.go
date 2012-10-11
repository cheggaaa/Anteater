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
	"crypto/md5"
	"os"
	"time"
	"cnst"
	"dump"
	"fmt"
	"stats"
	"aelog"
	"amazon"
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
	cf   *amazon.CloudFront
}


func GetStorage(c *config.Config) (s *Storage) {
	s = &Storage{
		Conf : c,
		wm   : &sync.Mutex{},
	}
	fmt.Printf("Try restore from %s... ", s.DumpFilename())
	err, exists := s.Restore()
	if err != nil {
		if  ! exists {
			fmt.Printf("index does not exists, create new storage.. ")
			s.Create()
		} else {
			panic(err)
		}
	}
	fmt.Print("done\n")
	s.Init()
	return
}


func (s *Storage) Init() {
	s.Stats = stats.New()
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
	if s.Conf.AmazonCFEnable {
		s.cf = amazon.NewCloudFront(s.Conf)
	}
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
			if f.CId != 0 {
				f.Delete()
			}
		}
	}()
	
	err := s.allocateFile(f)
	if err != nil {
		panic(err)
	}
	
	var written int64
	h := md5.New()
	buf := make([]byte, 32*1024)
	for {
		nr, er := r.Read(buf)
		fmt.Printf("Debug read: %v, %v\n", nr, er)
		if nr > 0 {
			nw, ew := f.WriteAt(buf[0:nr], written)
			fmt.Printf("Debug write: %v, %v\n", nw, ew)
			if nw > 0 {
				written += int64(nw)
				h.Write(buf[0:nw])
				s.Stats.Traffic.Input.AddN(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	
	if err != nil {
		panic(err)
	}
	
	if written != size {
		panic(fmt.Sprintf("Error while adding file. Requested size %d, but written only %d", size, written))
	}
	
	f.Md5 = h.Sum(nil)	
	
	err = s.Index.Add(name, f)
	if err != nil {
		panic(err)
	}
	s.Stats.Counters.Add.Add()
	ok = true
	return
}

func (s *Storage) allocateFile(f *File) (err error) {
	s.wm.Lock()
	defer s.wm.Unlock()
	var targets = []int{TARGET_SPACE_EQ, TARGET_SPACE_FREE, TARGET_NEW}
	var ok bool
	for _, target := range targets {
		for _, c := range s.Containers.Containers {
			if c.MaxSpace() >= f.Size {
				ok = c.Add(f, target)
				if ok {
					switch target {
					case TARGET_SPACE_EQ:
						s.Stats.Allocate.Replace.Add()
						break
					case TARGET_SPACE_FREE:
						s.Stats.Allocate.In.Add()
						break
					case TARGET_NEW:
						s.Stats.Allocate.Append.Add()
						break
					}
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
		if s.cf != nil {
			s.cf.OnChange(name)
		}
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
	aelog.Debugf("Dump: %s bytes writed to %s for %v prep(%v)", utils.HumanBytes(int64(n)), fname, tot, prep)
	return
}

func (s *Storage) Drop() (err error) {
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
	for _, c := range s.Containers.Containers {
		c.Close()
	}
}

func (s *Storage) DumpFilename() string {
	return s.Conf.DataPath + "index"
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
	
	for _, f := range s.Index.Files {
		s.Stats.Storage.FilesSize += f.Size
	}
		
	return s.Stats
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