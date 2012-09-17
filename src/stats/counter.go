package stats

import (
	"sync/atomic"
)

type Counter struct {
	c uint64
}

func (c *Counter) Add() uint64 {
	return atomic.AddUint64(&c.c, 1)
}

func (c *Counter) AddN(n int) uint64 {
	return atomic.AddUint64(&c.c, uint64(n))
}

func (c *Counter) GetValue() uint64 {
	return atomic.LoadUint64(&c.c)
}