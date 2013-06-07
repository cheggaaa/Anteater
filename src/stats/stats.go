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
	FilesCount   int64
	FilesSize    int64
	TotalSize    int64
	FreeSpace    int64
	HoleCount    int64
	HoleSize     int64
	IndexVersion int64
	DumpSize     int64
	DumpSaveTime time.Duration
	DumpLockTime time.Duration
	DumpTime     time.Time
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