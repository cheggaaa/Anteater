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