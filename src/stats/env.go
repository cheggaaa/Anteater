package stats

import (
	"time"
	"runtime"
)

type Env struct {
	GoVersion string
	NumGoroutine int
	Time time.Time
	MemAlloc uint64
}

func (e *Env) Refresh() {
	e.GoVersion = runtime.Version()
	e.NumGoroutine = runtime.NumGoroutine()
	e.Time = time.Now()
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	e.MemAlloc = m.Alloc
}