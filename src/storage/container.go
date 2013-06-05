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
	aelog.Debugln("try insert", f.Name)
	if s := c.holeIndex.GetBiggest(f.Index()); s != nil {
		aelog.Debugf("Found hole %v", s)
		if s.Index() == f.Index() {
			aelog.Debugln("Equal index hole %v, replace...", s.Index())
			f.SetOffset(s.Offset())
			c.replace(s, f)
			ok = true
			return
		}
		aelog.Debugf("Not equal index hole %d vs %d, normalize...", s.Index(), f.Index())
		
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
		h := c.insertNormalizedHole(s, s.Size()-f.Size())
		ok = true
		aelog.Debugln("Before normalize")
		c.Print()
		c.holeIndex.Print()		
		c.normalizeHole(h)
		aelog.Debugln("After normalize")
		fmt.Println(c.Check())
		c.Print()
		c.holeIndex.Print()
	}
	return
}

func (c *Container) Delete(f *File) {
	aelog.Debugln("c delete", f.Name)
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
	aelog.Debugln("Start normalize, find first from", h.Offset())
	i := 0
	for h.Prev() != nil && h.Prev().IsFree() {
		h = h.Prev().(*Hole)
		i++
		if i > 200 {
			break
		}
	}

	start := h.(*Hole)
	aelog.Debugln("First found", i, h.Offset())
	s := h.Size()
	h = h.Next()
	// check to right
	i = 0
	for {
		if ! h.IsFree() {
			break
		}
		s += h.Size()
		aelog.Debugf("Normalize: %d vs %d", R.Round(s), s)
		if R.Round(s) == s {
			start = c.mergeHoles(start, h.(*Hole))
			h = start
		}
		
		i++
		if h.Next() == nil || i > 400 {
			break
		}
		h = h.Next()
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
	aelog.Debugln("inh:", h.Indx, size)
	if R.Round(size) == size {
		h.Indx = R.Index(size)
		aelog.Debugln("inh:", "add!", h.Indx)
		c.holeIndex.Add(h)
		return h
	}	
	h.Indx = R.Index(size) - 1
	aelog.Debugln("inh:", "reduce", h.Indx, "next")
	if size == 1 {
		panic("Check algo!")
	}
	next := &Hole{
		Off:  h.End(),
		prev: h,
		next: h.Next(),
	}
	if next.next != nil {
		next.next.SetPrev(next)
	}
	h.SetNext(next)
	c.holeIndex.Add(h)
	c.insertNormalizedHole(next, size-h.Size())
	return h
}

func (c *Container) Print() {
	if c.last == nil {
		fmt.Println("EMPTY")
		return
	}
	
	fmt.Printf("C%d. Size: %s(%s); BI: %d\n", c.Id, utils.HumanBytes(c.Size), utils.HumanBytes(c.FileSize), c.holeIndex.biggestIndex)
	
	res := ""
	var s Space
	s = c.last
	i := 0
	for s != nil {
		i++
		n := "F"
		if s.IsFree() {
			n = "H"
			if ! c.holeIndex.Exists(s.(*Hole)) {
				n = "E"
			}
		}
		res = fmt.Sprintf("-%s%d-", n, s.Index()) + res
		s = s.Prev()
	}
	fmt.Printf("%s\n", res)
}

func (c *Container) Check() (err error) {
	var s,p Space
	s = c.last
	i := 0
	for s != nil {
		i++
		if s.IsFree() {
			if ! c.holeIndex.Exists(s.(*Hole)) {
				return fmt.Errorf("Hole not indexed: %v", s.(*Hole))
			}
		}
		
		if p != nil && s.End() != p.Offset() {
			return fmt.Errorf("Range error: %d vs %d", s.End(), p.Offset())
		}
		
		p = s
		s = s.Prev()
	}
	return
}
