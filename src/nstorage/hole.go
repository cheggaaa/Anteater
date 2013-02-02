package nstorage

type Hole struct {
	// Next and Prev
	Pr, Nt Space
	// Offset
	Off int64
	// Index
	Indx int
}

// implement Space
func (h *Hole) Next() Space {
	return h.Nt
}

func (h *Hole) Prev() Space {
	return h.Pr
}

func (h *Hole) SetNext(s Space) {
	h.Nt = s
}

func (h *Hole) SetPrev(s Space) {
	h.Pr = s
}

func (h *Hole) SetOffset(o int64) {
	h.Off = o
}

func (h *Hole) Offset() int64 {
	return h.Off
}

func (h *Hole) Size() int64 {
	return R.Size(h.Indx)
}

func (h *Hole) Index() int {
	return h.Indx
}

func (h *Hole) End() int64 {
	return h.Size() + h.Off
}

func (h *Hole) IsFree() bool {
	return true
}

func NewHole(offset, size int64) *Hole {
	return &Hole{Off: offset, Indx: R.Index(size)}
}
