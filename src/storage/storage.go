package storage

import (
	"io"
	"config"
	"sync"
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


func GetStorage(c *config.Config) (s *Storage, err error) {
	s = &Storage{
		Conf : c,
		wm   : &sync.Mutex{},
	}
	err = s.Init()
	return
}

func (s *Storage) Init() (err error) {
	return
}

func (s *Storage) Add(name string, r *io.Reader) {

}

func (s *Storage) Get(name string) (f *File, ok bool) {
	f, ok = s.Index.Get(name)
	return
}

func (s *Storage) Delete(name string) (ok bool) {
	return
}