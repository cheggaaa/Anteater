package storage

type Space interface {
	SetNext(Space)
	SetPrev(Space)
	Next() Space
	Prev() Space
	SetOffset(int64)
	Offset() int64
	Size() int64
	Index() int
	End() int64
	IsFree() bool
}
