package storage

import (
	"github.com/petar/GoLLRB/llrb"
)

var nullSpace = &Space{0, 0}

type Holes struct {
	Tree *llrb.Tree
	MaxSize int64
}


func (h *Holes) Init() {
	f := func(a, b interface{}) bool { return a.(*Space).Start < b.(*Space).Start }
	if h.Tree == nil {
		h.Tree = llrb.New(f)
	} else {
		h.Tree.Init(f)
	}
}

func (hs *Holes) Add(h *Space) {
	hs.Tree.ReplaceOrInsert(h)
}

func (hs *Holes) Allocate(size int64, target int) (start int64, found bool) {
	switch target {
	case TARGET_SPACE_EQ:
		hs.Tree.AscendGreaterOrEqual(nullSpace, func(i llrb.Item) bool {
			if i.(*Space).Size == size {
				hs.Tree.Delete(i)
				start = i.(*Space).Start
				return false
			}
			return true
		})
		return
	case TARGET_SPACE_FREE:
		var mx int64
		hs.Tree.AscendGreaterOrEqual(nullSpace, func(i llrb.Item) bool {
			if i.(*Space).Size >= size {
				i.(*Space).Size -= size 
				start = i.(*Space).Start + i.(*Space).Size
				return false
			}
			if i.(*Space).Size >= mx {
				mx = i.(*Space).Size
			}
			return true
		})
		if ! found { 
			hs.MaxSize = mx
		}
		return
	}
	return
}


func (hs *Holes) Clean(offset int64) (int64) {
	var prev *Space
	hs.Tree.AscendGreaterOrEqual(nullSpace, func(i llrb.Item) bool {
			h := i.(*Space)
			// is last
			if h.Start + h.Size == offset {
				offset -= h.Size
				h.Size = 0
			}
			
			// can merge	
			if prev != nil && (prev.Start + prev.Size) == h.Start {
				prev.Size += h.Size
				if hs.MaxSize < prev.Size {
					hs.MaxSize = prev.Size
				}
				h.Size = 0
			} else {
				prev = h
			}
			
			// need to delete
			if h.Size == 0 {
				hs.Tree.Delete(i)
			}
			return true
		})
	return offset
}

func (hs *Holes) DumpData() (s Spaces) {
	s = make(Spaces, hs.Tree.Len())
	i := 0
	hs.Tree.AscendGreaterOrEqual(nullSpace, func(h llrb.Item) bool {
		s[i] = h.(*Space)
		i++
		return true
	})
	return
}
