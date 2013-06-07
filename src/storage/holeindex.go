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
	//"aelog"
	"fmt"
)

type HoleIndex struct {
	// map[index][offset]*Hole
	index        map[int]map[int64]*Hole
	biggestIndex int
	Count        int64
	Size         int64
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

func (hi *HoleIndex) Exists(h *Hole) bool {
	if m, ok := hi.index[h.Index()]; ok {
		if h, ok := m[h.Offset()]; ok {
			return h == m[h.Offset()]
		}
	}
	return false
}

func (hi *HoleIndex) GetBiggest(index int) *Hole {
	if index > hi.biggestIndex {
		return nil
	}
	for ; index <= hi.biggestIndex; index++ {
		s := hi.Get(index)
		if s != nil {
			return s
		}
	}
	hi.reCalculateBiggestIndex()
	return nil
}

func (hi *HoleIndex) Add(holes... *Hole) {
	for _, s := range holes {
		index := s.Index()
		if index == 0 {
			panic("Check logick! Try to insert 0-index hole")
		}
		_, ok := hi.index[index]
		if !ok {
			hi.index[index] = make(map[int64]*Hole)
		}
		hi.index[index][s.Offset()] = s
	
		if index > hi.biggestIndex {
			hi.biggestIndex = index
		}
		hi.Count++
		hi.Size += s.Size()
	}
}

func (hi *HoleIndex) Create(prev, next Space, offset int64, index int) (h *Hole) {
	defer hi.Add(h)
	h = &Hole{
		prev: prev,
		next: next,
		Off:  offset,
		Indx: index,
	}
	return
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
