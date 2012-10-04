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

package storage

import (
	"sync"
	"sync/atomic"
)

type Containers struct {
	Containers map[int32]*Container
	LastId     int32
	m          *sync.Mutex
	s          *Storage
}

func (cs *Containers) Create() (c *Container, err error) {
	cs.m.Lock()
	defer cs.m.Unlock()
	id := atomic.AddInt32(&cs.LastId, 1)
	c = &Container{
		Id : id,
		s  : cs.s,
	}
	err = c.Create()
	cs.Containers[id] = c
	return
}

func (cs *Containers) Get(id int32) *Container {
	return cs.Containers[id]
}