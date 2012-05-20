package anteater

import (
	"time"
	"runtime"
	"sort"
	"sync/atomic"
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
	SpacesCount    int64
	SpacesSize     int64
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
		LastDumpTime : LastDumpTime,
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
	
	for _, id := range(ids) {
		cnt = FileContainers[int32(id)]
		_, sc, st := cnt.Spaces.ToHtml(10, cnt.Size)
		totalSize += cnt.Size
		fileCount += cnt.Count
		spacesCount += sc
		spacesTotalSize += st		
		allocated := cnt.Size - (cnt.Size - cnt.Offset) - st 
		totalFileSize += allocated
	}

	f := &StateFiles{
		ContainersCount : len(ids),
		FilesCount      : fileCount,
		FilesSize       : totalFileSize,
		SpacesCount     : spacesCount,
		SpacesSize      : spacesTotalSize,
	}
	
	
	return &State{
		Main     : m,
		Files    : f,
		Counters : HttpCn,
		RatesSinceStart : HttpCn.AsRate(int64(time.Now().Unix() - StartTime.Unix())),
		RatesLast5Seconds : GetHttpStateByPeriod(1).AsRate(int64(5)),
		RatesLastMinute : GetHttpStateByPeriod(12).AsRate(int64(60)),
		RatesLast5Minutes : GetHttpStateByPeriod(60).AsRate(int64(300)),
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
