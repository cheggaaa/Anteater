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
	"encoding/binary"
	"fmt"
	"github.com/cheggaaa/Anteater/aelog"
	"github.com/cheggaaa/Anteater/dump"
	"github.com/cheggaaa/Anteater/utils"
	"io"
	"os"
	"sync"
	"time"
)

const (
	ALLOC_REPLACE = 1
	ALLOC_APPEND  = 2
	ALLOC_INSERT  = 3
)

var Targets = []int{ALLOC_REPLACE, ALLOC_APPEND, ALLOC_INSERT}

type Container struct {
	Id                  int64
	Size                int64
	FileCount, FileSize int64
	FileRealSize        int64
	Created             bool
	last                *File
	holeIndex           *HoleIndex
	s                   *Storage
	f                   *os.File
	m                   *sync.Mutex
	ch                  bool
}

func (c *Container) Init(s *Storage, rr *dump.ResultReader) (err error) {
	aelog.Debugln("Init container", c.Id)
	c.m = new(sync.Mutex)
	c.s = s
	// open file
	c.f, err = os.OpenFile(c.fileName(), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	c.holeIndex = new(HoleIndex)
	c.holeIndex.Init(s.Conf.ContainerSize)
	if err = c.restore(rr); err != nil {
		return
	}
	c.ch = true
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
		err = c.Dump()
	}

	return
}

func (c *Container) create() (err error) {
	c.Size = c.s.Conf.ContainerSize
	if err = c.falloc(); err != nil {
		aelog.Infoln("Fallocate doesn't work:", err, "\nTry to truncate...")
		if err = c.fallocTruncate(); err != nil {
			return
		}
	}
	return
}

func (c *Container) fallocTruncate() (err error) {
	if err = c.f.Truncate(c.Size); err != nil {
		return
	}
	_, err = c.f.WriteAt([]byte{1}, c.Size-1)
	return
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

type dumper struct {
	c           *Container
	contFlushed bool
	next        Space
}

func (d *dumper) Write(w io.Writer) error {
	if !d.contFlushed {
		d.contFlushed = true
		if d.c.last != nil {
			d.next = d.c.last
		}
		return d.c.MarshallTo(w)
	}
	if d.next != nil {
		toReturn := d.next
		d.next = toReturn.Prev()
		return toReturn.MarshalTo(w)
	}
	return io.EOF
}

func (c *Container) Dump() (err error) {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.ch {
		return
	}

	st := time.Now()
	pr := time.Since(st)

	n, err := dump.DumpTo(c.indexName(), &dumper{c: c})
	aelog.Debugf("Dump container %d, writed %s for a %v (prep: %v)", c.Id, utils.HumanBytes(n), time.Since(st), pr)
	c.ch = false
	return
}

func (c *Container) restore(rr *dump.ResultReader) (err error) {
	var i int64
	if rr == nil {
		return nil
	}

	c.FileRealSize = 0
	var sp, prev Space
	sp, err = UnmarshallSpace(rr.B)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return
	}
	if sp != nil {
		c.last = sp.(*File)
	}
	if c.last != nil {
		c.last.Init(c)
		c.s.Index.Add(c.last)
		c.FileRealSize += c.last.Size()
		prev = c.last
	}
	i++

	for {
		if sp, err = UnmarshallSpace(rr.B); err != nil {
			if err == io.EOF {
				aelog.Debugf("read %d files", i)
				return nil
			}
			return
		}
		if lastF, ok := sp.(*File); ok {
			lastF.SetNext(prev)
			prev.SetPrev(lastF)
			lastF.Init(c)
			c.s.Index.Add(lastF)
			c.FileRealSize += lastF.Size()
		} else {
			lastS := sp.(*Hole)
			lastS.SetNext(prev)
			prev.SetPrev(lastS)
			c.holeIndex.Add(lastS)
		}
		prev = sp
		i++
	}
	return nil
}

// Allocator

func (c *Container) Allocate(f *File, target int) (ok bool) {
	c.m.Lock()
	defer func() {
		if ok {
			c.FileCount++
			c.FileSize += f.FSize
			c.FileRealSize += f.Size()
			f.Init(c)
			c.ch = true
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
	if h := c.holeIndex.Get(int(f.Index())); h != nil {
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
	if s := c.holeIndex.GetBiggest(int(f.Index())); s != nil {
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
		h := c.insertNormalizedHole(s, s.Size()-f.Size())
		ok = true
		c.normalizeHole(h)
	}
	return
}

func (c *Container) Delete(f *File) {
	c.m.Lock()
	defer c.m.Unlock()
	c.ch = true
	c.FileCount--
	c.FileSize -= f.FSize
	c.FileRealSize -= f.Size()
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
		Indx: int32(f.Index()),
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
	for {
		if !h.IsFree() {
			break
		}
		s += h.Size()
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
			if !c.holeIndex.Exists(s.(*Hole)) {
				n = "E"
			}
		}
		res = fmt.Sprintf("-%s%d-", n, s.Index()) + res
		s = s.Prev()
	}
	fmt.Printf("%s\n", res)
}

func (c *Container) Check() (err error) {
	if c.last == nil {
		return
	}
	var s, p Space
	s = c.last
	i := 0
	for s != nil {
		i++
		if s.IsFree() {
			if !c.holeIndex.Exists(s.(*Hole)) {
				return fmt.Errorf("Hole not indexed: %v", s.(*Hole))
			}
		} else {
			if e := s.(*File).CheckMd5(); e != nil {
				return e
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

func (c *Container) MarshallTo(w io.Writer) error {
	var arr [binary.MaxVarintLen64 * 6]byte
	buf := arr[:0]
	buf = append(buf, 3)
	buf = binary.AppendUvarint(buf, uint64(c.Id))
	buf = binary.AppendUvarint(buf, uint64(c.Size))
	buf = binary.AppendUvarint(buf, uint64(c.FileCount))
	buf = binary.AppendUvarint(buf, uint64(c.FileSize))
	buf = binary.AppendUvarint(buf, uint64(c.FileRealSize))
	if c.Created {
		buf = append(buf, 11)
	} else {
		buf = append(buf, 10)
	}
	_, err := w.Write(buf)
	return err
}
