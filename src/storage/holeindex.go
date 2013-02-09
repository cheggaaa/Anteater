package storage

import (
	"fmt"
)

type HoleIndex struct {
	// map[index][offset]*Hole
	index map[int]map[int64]*Hole
	biggestIndex int
	Count int64
	Size int64
}

func (hi *HoleIndex) Init() {
	hi.index = make(map[int]map[int64]*Hole)
}

func (hi *HoleIndex) Get(index int) *Hole {
	l, ok := hi.index[index]
	if ok {
		for _, s := range l {
			hi.Delete(s)
			return s
		}
	}
	return nil
}

func (hi *HoleIndex) GetBiggest(index int) *Hole {
	if index > hi.biggestIndex {
		return nil
	}
	for ;index <= hi.biggestIndex; index++ {
		s := hi.Get(index)
		if s != nil {
			return s
		}
	}
	hi.reCalculateBiggestIndex()
	return nil
}

func (hi *HoleIndex) Add(s *Hole) {
	index := s.Index()
	if index == 0 {
		panic("Check logick! Try to insert 0-index hole")
	}
	_, ok := hi.index[index]
	if ! ok {
		hi.index[index] = make(map[int64]*Hole)
	}
	hi.index[index][s.Offset()] = s
	
	if index > hi.biggestIndex {
		hi.biggestIndex = index
	}
	hi.Count++
	hi.Size += s.Size()
}

func (hi *HoleIndex) Delete(s *Hole) {
	if l, ok := hi.index[s.Index()]; ok {
		if _, ok = hi.index[s.Index()][s.Offset()]; ok {
			delete(l, s.Offset())
			hi.Count--
			hi.Size -= s.Size()
			if len(l) == 0 && s.Index() == hi.biggestIndex {
				hi.reCalculateBiggestIndex()
			}
		}
	}
}

func (hi *HoleIndex) reCalculateBiggestIndex() {
	hi.biggestIndex = 0
	for i, d := range hi.index {
		if i > hi.biggestIndex && len(d) > 0 {
			hi.biggestIndex = i
		}
	}
}

func (hi *HoleIndex) Print() {
	for _, l := range hi.index {
		for _, s := range l {
			fmt.Printf("%d\t%d\t%d\t%d\n", s.Offset(), s.Index(), s.Size(), s.End())
		}
	}
}