package stats

import (
	"storage"
	"time"
	"cnst"
)

type Stats struct {
	Anteater *Anteater
	Storage *Storage
	Allocate *Allocate
	Counters *StorageCounters
	Http *Http
	Env *Env
	s *storage.Storage
}

type Allocate struct {
	Append, Replace, In *Counter
}

type Anteater struct {
	Version string
	StartTime time.Time
}

type StorageCounters struct {
	Get, Add, Delete, NotFound *Counter
}

type Http struct {
	Input, Output *Counter
}

func New(s *storage.Storage) *Stats {
	st := &Stats{}
	
	st.Anteater = &Anteater{
		Version : cnst.VERSION,
		StartTime : time.Now(),
	}
	
	st.Storage = &Storage{}
	st.Storage.Refresh(s)
	
	st.Allocate = &Allocate{&Counter{}, &Counter{}, &Counter{}}
	st.Counters = &StorageCounters{&Counter{}, &Counter{}, &Counter{}, &Counter{}}
	st.Http = &Http{&Counter{}, &Counter{}}
	st.Env = &Env{}
	st.Env.Refresh()
	
	return st
}


func (s *Stats) Refresh() {
	s.Env.Refresh()
	s.Storage.Refresh(s.s)
}