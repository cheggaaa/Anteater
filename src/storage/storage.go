package storage

import (
	"io"
	"config"
	"sync"
	"errors"
	"crypto/md5"
)

const (
	TARGET_SPACE_EQ   = 0
	TARGET_SPACE_FREE = 1
	TARGET_NEW        = 2
)

type Storage struct {
	Index *Index
	Containers *Containers
	Conf *config.Config
	wm   *sync.Mutex
}


func GetStorage(c *config.Config) (s *Storage) {
	s = &Storage{
		Conf : c,
		wm   : &sync.Mutex{},
	}
	return
}


func (s *Storage) Init() (err error) {
	return
}

func (s *Storage) Create() (err error) {
	s.Index = &Index{make(map[string]*File), &sync.Mutex{}}
	s.Containers = &Containers{
		s : s,
		m : &sync.Mutex{},
		Containers : make(map[int32]*Container),
	}
	_, err = s.Containers.Create()
	return
}

func (s *Storage) Add(name string, r io.Reader, size int64) (f *File) {
	f = &File{
		Size: size,
		s : s,
	}
	var ok bool
	defer func() {
		if ! ok {
			if f.CId != 0 {
				f.c.Delete(f)
			}
		}
	}()
	
	err := s.allocateFile(f)
	if err != nil {
		panic(err)
	}
	
	var written int64
	h := md5.New()
	for {
		buf := make([]byte, 100*1024)
		nr, er := io.ReadFull(r, buf)
		if nr > 0 {
			nw, ew := f.WriteAt(buf[0:nr], written)
			if nw > 0 {
				written += int64(nw)
				h.Write(buf[0:nw])
			}
			if ew != nil {
				err = ew
				break
			}
		}
		if er != nil {
			err = er
			break
		}
		if err != nil {
			panic(err)
		}
	}
	
	f.Md5 = h.Sum(nil)	
	
	s.Index.Add(name, f)
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

func (s *Storage) Close() {
	for _, c := range s.Containers.Containers {
		c.Close()
	}
}