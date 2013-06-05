package storage

import (
	"testing"
)

func TestRound(t *testing.T) {
	c := 1000
	ls := int64(0)
	for i := 1; i <= c; i++ {
		size := R.Size(i)
		indx := R.Index(size)
		t.Logf("I%d\t%d", indx, size)
		if indx != i {
			t.Errorf("Size-index conversion failed! %d vs %d", indx, i)
		}
		if size <= ls {
			t.Errorf("Last size more then actual: %d vs %d", ls, size)
		}
		ls = size
	}
} 