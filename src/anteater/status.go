package anteater

import (
	"time"
	"runtime"
	"sort"
	"sync/atomic"
)

type State struct {
	Main  *StateMain
	Files *StateFiles
	Counters *StateHttpCounters
	Alloc *StateAllocateCounters
}

type StateMain struct {
	Time          int64
	StartTime        int64
	Goroutines    int
	MemoryUsage   uint64
	Requests      map[string]int64
	RequestsRate  map[string]int64
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



func GetState() *State {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	m := &StateMain{
		Time         : time.Now().Unix(),
		StartTime    : StartTime.Unix(),
		Goroutines   : runtime.NumGoroutine(),
		MemoryUsage  : ms.TotalAlloc,
		Requests     : make(map[string]int64),
		RequestsRate : make(map[string]int64),
		LastDump     : LastDump.Unix(),
		LastDumpTime : LastDumpTime,
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
		Alloc    : AllocCn,
	}
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
