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
	"time"
	"runtime"
	"sort"
	"sync/atomic"
	"encoding/json"
)


var FiveSecondsCounters []*StateHttpCounters = make([]*StateHttpCounters, 61)
var FiveSecondsCursor int

type State struct {
	Main  *StateMain
	Files *StateFiles
	Counters *StateHttpCounters
	RatesSinceStart   *StateHttpCounters
	RatesLast5Seconds *StateHttpCounters
	RatesLastMinute   *StateHttpCounters
	RatesLast5Minutes *StateHttpCounters
	Alloc *StateAllocateCounters
}

type StateMain struct {
	Time          int64
	StartTime     int64
	Goroutines    int
	LastDump      int64
	LastDumpTime  int64
	IndexFileSize int64
}

type StateFiles struct {
	ContainersCount int
	FilesCount      int64
	FilesSize       int64	
	SpacesCount    	int64
	SpacesSize     	int64
	ByContainers   	[][]int64
}

type StateHttpCounters struct {
	Get      int64
	Add      int64
	Delete   int64
	NotFound int64
}

type StateAllocateCounters struct {
	ReplaceSpace int64
	AddToSpace   int64
	ToEnd        int64
}


func init() {
	FiveSecondsTick()
	go func() { 
		ch := time.Tick(5 * time.Second)
		for _ = range ch {
			func () {
				FiveSecondsTick()
			}()
		}
	}()
}


func GetState() *State {
	m := &StateMain{
		Time         : time.Now().Unix(),
		StartTime    : StartTime.Unix(),
		Goroutines   : runtime.NumGoroutine(),
		LastDump     : LastDump.Unix(),
		LastDumpTime : LastDumpTime.Nanoseconds(),
		IndexFileSize: IndexFileSize,
	}
	
	GetFileLock.Lock()
	defer GetFileLock.Unlock()

	var ids []int = make([]int, 0)
	for _, cn := range(FileContainers) {
		ids = append(ids, int(cn.Id))
	}
	sort.Ints(ids)
	var cnt *Container
	
	var totalSize, totalFileSize, fileCount, spacesCount, spacesTotalSize int64
	byCont := make([][]int64, len(ids))
	
	i := 0
	for _, id := range(ids) {
		cnt = FileContainers[int32(id)]
		sc, st := cnt.Spaces.Stats()
		totalSize += cnt.Size
		fileCount += cnt.Count
		spacesCount += sc
		spacesTotalSize += st		
		allocated := cnt.Size - (cnt.Size - cnt.Offset) - st 
		totalFileSize += allocated
		byCont[i] = []int64{cnt.Size, allocated, st}
		i++
	}

	f := &StateFiles{
		ContainersCount : len(ids),
		FilesCount      : fileCount,
		FilesSize       : totalFileSize,
		SpacesCount     : spacesCount,
		SpacesSize      : spacesTotalSize,
		ByContainers    : byCont,
	}
	
	rateStart := time.Now().Unix() - StartTime.Unix()
	
	rate1 := int64(5)
	if time.Now().Unix() - rate1 < StartTime.Unix() {
		rate1 = rateStart
	}
	
	rate2 := int64(60)
	if time.Now().Unix() - rate2 < StartTime.Unix() {
		rate2 = rateStart
	}
	
	rate3 := int64(300)
	if time.Now().Unix() - rate3 < StartTime.Unix() {
		rate3 = rateStart
	}
	
	
	return &State{
		Main     : m,
		Files    : f,
		Counters : HttpCn,
		RatesSinceStart : HttpCn.AsRate(rateStart),
		RatesLast5Seconds : GetHttpStateByPeriod(1).AsRate(rate1),
		RatesLastMinute : GetHttpStateByPeriod(12).AsRate(rate2),
		RatesLast5Minutes : GetHttpStateByPeriod(60).AsRate(rate3),
		Alloc    : AllocCn,
	}
}


func FiveSecondsTick() {
	if FiveSecondsCursor == 61 {
		FiveSecondsCursor = 0
	}
	if FiveSecondsCounters[FiveSecondsCursor] == nil {
		FiveSecondsCounters[FiveSecondsCursor] = &StateHttpCounters{}
	}
	FiveSecondsCounters[FiveSecondsCursor].SetData(HttpCn)	
	FiveSecondsCursor++
}

func GetHttpStateByPeriod(period int) (result *StateHttpCounters) {
	curCursor := FiveSecondsCursor - 1
	diffCursor := curCursor
	result = &StateHttpCounters{}
	i := 0
	for {
		diffCursor--
		i++
		if diffCursor == -1 {
			diffCursor = 60
		}		
		if FiveSecondsCounters[diffCursor] == nil {
			diffCursor++
			if diffCursor == 61 {
				diffCursor = 0
			} 
			break
		}
		if i >= period {
			break
		}
	} 

	cur  := FiveSecondsCounters[curCursor]
	diff := FiveSecondsCounters[diffCursor]
	result.Add = cur.Add - diff.Add
	result.Get = cur.Get - diff.Get
	result.Delete = cur.Delete - diff.Delete
	result.NotFound = cur.NotFound - diff.NotFound
	return
}

func (s State) AsJson() (b []byte) {
	b, err := json.Marshal(s)	
	if err != nil {
		Log.Warnln(err)
	}
	return
}


func (s *StateHttpCounters) CGet() {
	atomic.AddInt64(&s.Get, 1)
}

func (s *StateHttpCounters) CAdd() {
	atomic.AddInt64(&s.Add, 1)
}

func (s *StateHttpCounters) CDelete() {
	atomic.AddInt64(&s.Delete, 1)
}

func (s *StateHttpCounters) CNotFound() {
	atomic.AddInt64(&s.NotFound, 1)
}

func (s *StateHttpCounters) SetData(otherS *StateHttpCounters) {
	s.Get = otherS.Get
	s.Add = otherS.Add
	s.Delete = otherS.Delete
	s.NotFound = otherS.NotFound
}

func (s *StateHttpCounters) Sum() int64 {
	return s.Get + s.Add + s.Delete + s.NotFound
}

func (s *StateHttpCounters) AsRate(duration int64) *StateHttpCounters {
	return &StateHttpCounters{
		Add : s.Add / duration,
		Get : s.Get / duration,
		Delete : s.Delete / duration,
		NotFound : s.NotFound / duration, 
	}
}

func (s *StateAllocateCounters) CTarget(target int) {
	switch target {
		case TARGET_SPACE_EQ:
			atomic.AddInt64(&s.ReplaceSpace, 1)
		case TARGET_SPACE_FREE:
			atomic.AddInt64(&s.AddToSpace, 1)
	 	case TARGET_NEW:
	 		atomic.AddInt64(&s.ToEnd, 1)
	}
}
