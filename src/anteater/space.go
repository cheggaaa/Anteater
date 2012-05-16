package anteater

import (
	"sort"
	"strconv"
	"errors"
)

type Space struct {
	Start int64
	Size  int64
}

type Spaces []*Space

/**
 * Remove all items with start in rem
 * Return new Spaces
 */
func (s Spaces) FilterByStart(rem []int64) Spaces {
	n := make(Spaces, 0)
	var add bool
	for _, space := range s {
		add = true
		for _, r := range rem {
			if r == space.Start {
				add = false
				break
			}
		}
		if add {
			n = append(n, space)
		}
	}
	n.Sort()
	return n
}

func (s Spaces) Join() (Spaces, int64) {
	var prev *Space
	var f bool = true
	var maxSpaceSize int64
	rem := make([]int64, 0)
	for _, space := range(s) {
		if space.Size == 0 {
			rem = append(rem, space.Start)
			continue
		}
		
		if maxSpaceSize < space.Size {
			maxSpaceSize = space.Size
		}
		
		if !f && (prev.Start + prev.Size) == space.Start {
			prev.Size += space.Size
			rem = append(rem, space.Start)
			if maxSpaceSize < prev.Size {
				maxSpaceSize = prev.Size
			}
		} else {
			prev = space
		}
		f = false	
	}
	return s.FilterByStart(rem), maxSpaceSize
}

func (s Spaces) Get(size int64, target int) (int64, error) {
	switch target {
	case TARGET_SPACE_EQ:
		for i, space := range(s) {
			if space.Size == size {			
				start := space.Start
				s[i].Size = 0
				return start, nil
			}
		}
	case TARGET_SPACE_FREE:
		for i, space := range(s) {
			if space.Size >= size {			
				start := space.Start
				s[i].Start += size
				s[i].Size  -= size
				return start, nil
			}
		}
	}
	return 0, errors.New("Can't allocate space")
}

// Sort interface

func (s Spaces) Len() int {
	return len(s)
}

func (s Spaces) Less(i, j int) bool {
	return s[i].Start < s[j].Start
}

func (s Spaces) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Spaces) Sort() {
	sort.Sort(s)
}

func (s Spaces) ToHtml(limit int, cSize int64) string {
	var total int64
	lhtml := "<div class=\"list\"><ul>"
	i := 0
	for _, space := range(s) {		
		if i < limit {
			lhtml += "<li>" + strconv.FormatInt(space.Start, 10) + ":" + strconv.FormatInt(space.Start + space.Size, 10) + "(" + strconv.FormatInt(space.Size, 10) + ")</li>"
		}
		total += space.Size
		i++
	}
	lhtml += "</div>"
	ihtml := "<div class=\"info\">"
	prc := (total * 100) / cSize 
	ihtml += "Len: <b>" + strconv.FormatInt(int64(len(s)), 10) + " (" + strconv.FormatInt(total, 10) + " bytes) " + strconv.FormatInt(prc, 10) + "%</b>"
	ihtml += "</div>"
	return  "<div class=\"spaces\">" + ihtml + lhtml + "</ul></div></div>"
} 
