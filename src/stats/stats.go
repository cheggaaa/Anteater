package stats

import (
	"time"
	"cnst"
)

type Stats struct {
	Anteater *Anteater
	Storage *Storage
	Allocate *Allocate
	Counters *StorageCounters
	Traffic *Traffic
	Env *Env
}


type Allocate struct {
	Append, Replace, In *Counter
}

type Anteater struct {
	Version string
	StartTime time.Time
}

type StorageCounters struct {
	Get, Add, Delete, NotFound, NotModified *Counter
}

type Traffic struct {
	Input, Output *Counter
}

type Storage struct {
	ContainersCount int
	FilesCount int
	FilesSize int64
	TotalSize int64
	FreeSpace int64
	HoleCount int
	HoleSize int64
	IndexVersion uint64
}

func New() *Stats {
	st := &Stats{}
	
	st.Anteater = &Anteater{
		Version : cnst.VERSION,
		StartTime : time.Now(),
	}
	
	st.Storage = &Storage{}
	
	st.Allocate = &Allocate{&Counter{}, &Counter{}, &Counter{}}
	st.Counters = &StorageCounters{&Counter{}, &Counter{}, &Counter{}, &Counter{}, &Counter{}}
	st.Traffic = &Traffic{&Counter{}, &Counter{}}
	st.Env = &Env{}
	st.Env.Refresh()
	
	return st
}


func (s *Stats) Refresh() {
	s.Env.Refresh()
}