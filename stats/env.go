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