package storage

import (
	"aelog"
	"dump"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
	"utils"
)

const (
	ALLOC_REPLACE = 1
	ALLOC_APPEND  = 2
	ALLOC_INSERT  = 3
)

var Targets = []int{ALLOC_REPLACE, ALLOC_APPEND, ALLOC_INSERT}

type Container struct {
	Id                  int64
	Created             bool
	Size                int64
	FileCount, FileSize int64
	Last                *File
	holeIndex           *HoleIndex
	s                   *Storage
	f                   *os.File
	m                   *sync.Mutex
}

func (c *Container) Init(s *Storage) (err error) {
	aelog.Debugln("Init container", c.Id)
	c.m = new(sync.Mutex)
	c.s = s
	// open file
	c.f, err = os.OpenFile(c.fileName(), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	c.holeIndex = new(HoleIndex)
	c.holeIndex.Init()

	// build holeIndex
	var last, next Space
	next = nil
	last = c.Last
	for last != nil && c.Last != nil {
		if last.IsFree() {
			c.holeIndex.Add(last.(*Hole))
			aelog.Debug("init hole")
		} else {
			f := last.(*File)
			f.Init(c)
			c.s.Index.Add(f)
		}
		if next != nil {
			last.SetNext(next)
		}
		next = last
		last = last.Prev()
	}

	// create
	if !c.Created {
		aelog.Debugln("Create conatiner", c.Id)
		if err = c.create(); err == nil {
			c.Created = true
		}
	}
	
	return
}

func (c *Container) create() (err error) {
	c.Size = c.s.Conf.ContainerSize
	if err = c.falloc(); err != nil {
		return
	}
	err = c.Dump()
	return
}

func (c *Container) falloc() (err error) {
	return syscall.Fallocate(int(c.f.Fd()), 0, 0, c.Size)
}

func (c *Container) fileName() string {
	return fmt.Sprintf("%sc%d.data", c.s.Conf.DataPath, c.Id)
}

func (c *Container) indexName() string {
	return fmt.Sprintf("%sc%d.index", c.s.Conf.DataPath, c.Id)
}

func (c *Container) Close() (err error) {
	return c.f.Close()
}

func (c *Container) Dump() (err error) {
	c.m.Lock()
	defer c.m.Unlock()
	st := time.Now()
	n, err := dump.DumpTo(c.indexName(), c)
	aelog.Debugf("Dump container %d, writed %s for a %v", c.Id, utils.HumanBytes(n), time.Since(st))
	return
}

// Allocator

func (c *Container) Allocate(f *File, target int) (ok bool) {
	c.m.Lock()
	defer func() {
		if ok {
			c.FileCount++
			c.FileSize += f.FSize
			f.c = c
		}
		c.m.Unlock()
	}()

	if f.Indx == 0 {
		f.Indx = R.Index(f.FSize)
	}

	// if first
	if c.Last == nil {
		c.Last = f
		ok = true
		return
	}

	switch target {
	case ALLOC_REPLACE:
		ok = c.allocReplace(f)
	case ALLOC_APPEND:
		ok = c.allocAppend(f)
	case ALLOC_INSERT:
		ok = c.allocInsert(f)
	}
	return
}

func (c *Container) allocReplace(f *File) (ok bool) {
	if h := c.holeIndex.Get(f.Index()); h != nil {
		f.SetOffset(h.Offset())
		c.replace(h, f)
		c.holeIndex.Delete(h)
		ok = true
	}
	return
}

func (c *Container) allocAppend(f *File) (ok bool) {
	if c.Size-c.Last.End() >= f.Size() {
		f.SetOffset(c.Last.End())
		f.SetPrev(c.Last)
		c.Last.SetNext(f)
		c.Last = f
		ok = true
	}
	return
}

func (c *Container) allocInsert(f *File) (ok bool) {
	if s := c.holeIndex.GetBiggest(f.Index()); s != nil {
		// insert to begin of hole
		f.SetPrev(s.Prev())
		f.SetNext(s)
		s.SetPrev(f)
		s.SetOffset(s.Offset() + f.Size())
		p := f.Prev()
		if p != nil {
			p.SetNext(f)
		}
		c.holeIndex.Delete(s)
		c.insertNormalizedHole(s, s.Size()-f.Size())
		ok = true
	}
	return
}

func (c *Container) Delete(f *File) {
	c.FileCount--
	c.FileSize -= f.FSize

	// is last
	if c.Last.Offset() == f.Offset() {
		prev := c.Last.Prev()

		// remove tail holes
		for prev != nil && prev.IsFree() {
			c.holeIndex.Delete(prev.(*Hole))
			prev = prev.Prev()
		}

		// is nothing
		if prev == nil {
			c.Last = nil
		} else {
			c.Last = prev.(*File)
			prev.SetNext(nil)
		}
		return
	}

	// create hole
	h := &Hole{
		Indx: f.Index(),
	}
	// replace file to hole
	h.SetOffset(f.Offset())
	c.replace(f, h)
	// add hole to index
	c.holeIndex.Add(h)
	// normalize
	c.normalizeHole(h)
}

func (c *Container) normalizeHole(h Space) {
	// find start hole
	i := 0
	for h.Prev() != nil && h.Prev().IsFree() {
		h = h.Prev().(*Hole)
		i++
		if i > 200 {
			break
		}
	}

	start := h.(*Hole)
	s := h.Size()
	h = h.Next()
	// check to right
	i = 0
	for h != nil && h.IsFree() {
		s += h.Size()
		if R.Round(s) == s {
			start = c.mergeHoles(start, h.(*Hole))
			h = start
		}
		h = h.Next()
		i++
		if i > 400 {
			break
		}
	}
}

// merge spaces from start to end, and return new space
func (c *Container) mergeHoles(start, end *Hole) (newHole *Hole) {
	// create hole
	newHole = &Hole{}
	newSize := end.End() - start.Offset()
	if R.Round(newSize) != newSize {
		panic("merge algo error")
	}

	// remove old
	rm := start
	for rm != end {
		c.holeIndex.Delete(rm)
		rm = rm.Next().(*Hole)
	}
	c.holeIndex.Delete(end)

	newHole.Indx = R.Index(newSize)
	newHole.SetOffset(start.Offset())
	c.replace(start, newHole)
	next := end.Next()
	newHole.SetNext(next)
	if next != nil {
		next.SetPrev(newHole)
	}
	// add to index
	c.holeIndex.Add(newHole)
	return
}

// Replace s1 to s2. Move offset, next and prev
func (c *Container) replace(s1, s2 Space) {
	s2.SetNext(s1.Next())
	s2.SetPrev(s1.Prev())
	n := s2.Next()
	if n != nil {
		n.SetPrev(s2)
	}
	p := s2.Prev()
	if p != nil {
		p.SetNext(s2)
		// check algo
		if p.End() != s2.Offset() {
			panic("Illegal offset!!!")
		}
	}
}

func (c *Container) insertNormalizedHole(h *Hole, size int64) *Hole {
	if R.Round(size) == size {
		h.Indx = R.Index(size)
		c.holeIndex.Add(h)
		return h
	}
	h.Indx = R.Index(size) - 1
	if size == 1 {
		panic("Check algo!")
	}
	next := &Hole{
		Off:  h.End(),
		pr:   h,
		next: h.Next(),
	}
	h.SetNext(c.insertNormalizedHole(next, size-h.Size()))
	return h
}
