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
	FDump 				map[int64]*File
	HDump               map[int64]*Hole
	last                *File
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
	c.restore()
	// build holeIndex
	/*var last, next Space
	next = nil
	last = c.last
	for last != nil && c.last != nil {
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
	*/

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
	
	var i int64
	
	c.FDump = make(map[int64]*File, c.FileCount)
	c.HDump = make(map[int64]*Hole, c.holeIndex.Count)
	
	var last Space = c.last
	
	for last != nil && c.last != nil {
		if last.IsFree() {
			c.HDump[i] = last.(*Hole)
		} else {
			c.FDump[i] = last.(*File)
		}
		last = last.Prev()
		i++
	}
	
	pr := time.Since(st)
	
	n, err := dump.DumpTo(c.indexName(), c)
	aelog.Debugf("Dump container %d, writed %s for a %v (prep: %v)", c.Id, utils.HumanBytes(n), time.Since(st), pr)
	return
}

func (c *Container) restore() {
	var i int64
	if c.FDump == nil || len(c.FDump) == 0 {
		return
	}
	
	c.last = c.FDump[i]
	var next Space = c.last
	i++
	
	for {
		if last, ok := c.FDump[i]; ok {
			last.SetNext(next)
			next.SetPrev(last)
			next = last
			last.Init(c)
			c.s.Index.Add(last)
		} else if last, ok := c.HDump[i]; ok {
			last.SetNext(next)
			next.SetPrev(last)
			next = last
			c.holeIndex.Add(last)
		} else {
			break
		}
		i++
	}
	c.FDump = nil
	c.HDump = nil
}

// Allocator

func (c *Container) Allocate(f *File, target int) (ok bool) {
	c.m.Lock()
	defer func() {
		if ok {
			c.FileCount++
			c.FileSize += f.FSize
			f.Init(c)
		}
		c.m.Unlock()
	}()

	if f.Indx == 0 {
		f.Indx = R.Index(f.FSize)
	}

	// if first
	if c.last == nil {
		c.last = f
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
		ok = true
	}
	return
}

func (c *Container) allocAppend(f *File) (ok bool) {
	if c.Size-c.last.End() >= f.Size() {
		f.SetOffset(c.last.End())
		f.SetPrev(c.last)
		c.last.SetNext(f)
		c.last = f
		ok = true
	}
	return
}

func (c *Container) allocInsert(f *File) (ok bool) {
	if s := c.holeIndex.GetBiggest(f.Index()); s != nil {
		if s.Index() == f.Index() {
			f.SetOffset(s.Offset())
			c.replace(s, f)
			ok = true
			return
		}
		// insert to begin of hole
		f.SetPrev(s.Prev())
		f.SetNext(s)
		f.SetOffset(s.Offset())
		s.SetPrev(f)
		s.SetOffset(s.Offset() + f.Size())
		p := f.Prev()
		if p != nil {
			p.SetNext(f)
		}
		c.insertNormalizedHole(s, s.Size()-f.Size())
		ok = true
	}
	return
}

func (c *Container) Delete(f *File) {
	c.m.Lock()
	defer c.m.Unlock()
	
	c.FileCount--
	c.FileSize -= f.FSize

	// is last
	if c.last.Offset() == f.Offset() {
		prev := c.last.Prev()

		// remove tail holes
		for prev != nil && prev.IsFree() {
			c.holeIndex.Delete(prev.(*Hole))
			prev = prev.Prev()
		}

		// is nothing
		if prev == nil {
			c.last = nil
		} else {
			c.last = prev.(*File)
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

// Replace s1 to s2. Move next and prev
func (c *Container) replace(s1, s2 Space) {
	n := s1.Next()
	s2.SetNext(n)
	if n != nil {
		n.SetPrev(s2)
	}
	p := s1.Prev()
	s2.SetPrev(p)
	if p != nil {
		p.SetNext(s2)
		// check algo
		if p.End() != s2.Offset() {
			panic(fmt.Sprintf("Illegal offset: %d vs %d\n%+v\n%+v", p.End(), s2.Offset(), s1, s2))
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
		prev: h,
		next: h.Next(),
	}
	h.SetNext(c.insertNormalizedHole(next, size-h.Size()))
	return h
}

func (c *Container) Print() {
	if c.last == nil {
		fmt.Println("EMPTY")
		return
	}
	
	var s Space
	s = c.last
	i := 0
	for s != nil {
		i++
		n := "F"
		if s.IsFree() {
			n = "H"
		}
		fmt.Printf("%s (%d)\t%d(%d)\t%d\t%d\n", n, i, s.Size(), s.Index(), s.Offset(), s.End())
		s = s.Prev()
	}
}
