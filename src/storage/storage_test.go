package storage

import (
	"testing"
	"crypto/rand"
	"io"
	"config"
	"utils"
	"crypto/md5"
	mrand "math/rand"
	"fmt"
	"sync"
)

var s *Storage

var fsize int64 = 1000
var it int64 = 100
var randFiles map[string]int64

func TestCreate(t *testing.T) {
	conf := storageConf()
	s = GetStorage(conf)
	s.Create()
	if len(s.Containers.Containers) != 1 {
		t.Fatalf("Wrong container count: %d", len(s.Containers.Containers))
	}
}

func TestAddSerial(t *testing.T) {
	var i int64
	for i = 0; i < it; i++ {
		addAndAssert(t, fmt.Sprintf("f-%d", i), fsize)
	}
	
	c := s.Containers.Get(1)
	
	if c.Count != int64(it) {
		t.Errorf("Files count mismatched! Expected:%d, actual:%d", it, c.Count)
	}
	
	if !c.ch {
		t.Errorf("Container must has status change")
	}
	
	if c.Offset != fsize * it {
		t.Errorf("Mismatch container offset! Expected: %d, actual: %d", fsize * it, c.Offset)
	}
}

func TestDeleteSerial(t *testing.T) {
	c := s.Containers.Get(1)
	var i int64
	for i = it - 1; i >= 0; i-- {
		n := fmt.Sprintf("f-%d", i)
		t.Logf("Storage.Delete(%s)", n)
		if ! s.Delete(n) {
			t.Errorf("Storage.Delete(%s) return false", n)
		}
		if c.Count != i {
			t.Errorf("Mismatch container count! Expected: %d, actual: %d", c.Count, i)
		}
		if c.Offset != c.Count * fsize {
			t.Errorf("Mismatch container offset! Expected: %d, actual: %d", c.Offset, c.Count * fsize)
		}
		if ! c.ch {
			t.Errorf("Container must has status change")
		}
	}
}

func TestAddRandom(t *testing.T) {
	createRandFiles(int(it))
	var totalSize int64
	for n, sz := range randFiles {
		addAndAssert(t, n, sz)
		totalSize += sz
	}
	c := s.Containers.Get(1)
	
	if c.Count != int64(it) {
		t.Errorf("Files count mismatched! Expected:%d, actual:%d", it, c.Count)
	}
	
	if !c.ch {
		t.Errorf("Container must has status change")
	}
	
	if c.Offset != totalSize {
		t.Errorf("Mismatch container offset! Expected: %d, actual: %d", totalSize, c.Offset)
	}
	
	addAndAssert(t, "last", 100)
}

func TestDeleteRandom(t *testing.T) {
	c := s.Containers.Get(1)
	for n, sz := range randFiles {
		if ! s.Delete(n) {
			t.Errorf("Storage.Delete(%s) return false", n)
		}
		if len(c.Spaces) != 1 {
			t.Errorf("Must be 1 space, but len(Spaces) %d", len(c.Spaces))
		}
		if c.Spaces[0].Size != sz {
			t.Errorf("Deleted file size (%d) and space size (%s) mismatched!", sz, c.Spaces[0].Size)
		} 
		if s.Delete(n) {
			t.Errorf("Storage.Delete(%s) return true, but file already deleted", n)
		}
		delete(randFiles, n)
		break
	}
	
	for n, _ := range randFiles {
		if ! s.Delete(n) {
			t.Errorf("Storage.Delete(%s) return false", n)
		}
		t.Logf("%v", c.Spaces[0])
	}
	if ! s.Delete("last") {
		t.Errorf("Storage.Delete(%s) return false", "last")
	}
	
	if c.Count != 0 {
		t.Errorf("Container has files: %d", c.Count)
	}
	
	if c.Offset != 0 {
		t.Errorf("Mismatch container offset! Expected: %d, actual: %d", 0, c.Offset)
	}
	
	if len(c.Spaces) != 0 {
		t.Errorf("Container has spaces: %d %v", len(c.Spaces), c.Spaces[0])
	}
}


func TestCreateContainer(t *testing.T) {
	for i := 0; i < 3; i++ {
		addAndAssert(t, fmt.Sprintf("f%d", i), 1024 * 1024)
	}
	if len(s.Containers.Containers) != 1 {
		t.Errorf("Must be 1 container, expected:%d", len(s.Containers.Containers))
	}
	c := s.Containers.Get(1)
	if c.Offset != c.Size {
		t.Errorf("Offset (%d) must be equals size (%d)", c.Offset, c.Size)
	}
	
	addAndAssert(t, "new", 1024 * 1024)
	
	if len(s.Containers.Containers) != 2 {
		t.Errorf("Must be 2 containers, expected:%d", len(s.Containers.Containers))
	}
	
	c = s.Containers.Get(2)
	
	if c == nil {
		t.Errorf("New Container has invalid id")
	}
	
	if c.Offset != 1024 * 1024 {
		t.Errorf("Invalid offset: %d, must be %d", c.Offset, 1024 * 1024)
	}
}

func TestAtomic(t *testing.T) {
	w := func(f map[string]int64, c int) {
		for i := 0; i < c; i++ {
			fc := f
			for n, sz := range fc {
				addAndAssert(t, n, sz)
				if mrand.Intn(2) > 1  {
					if ! s.Delete(n) {
						t.Errorf("Storage.Delete(%s) return false", n)
					}
					delete(fc, n)
				} 
			}
			for n, _ := range fc {
				f, ok := s.Get(n)
				if ! ok {
					t.Errorf("Storage.Get(%s) return false", n)
				}
				
				h := md5.New()
				io.Copy(h, f.GetReader())
				
				act := fmt.Sprintf("%x", h.Sum(nil))
				exp := fmt.Sprintf("%x", f.Md5)
				
				if act != exp {
					t.Error("Md5 file %s mismatch. Actual: %s, expected: %s", n, act, exp)
				}
				
				if ! s.Delete(n) {
					t.Errorf("Storage.Delete(%s) return false", n)
				}
			}
		}
	}
	
	wg := &sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		createRandFiles(400)
		f := randFiles
		go func (f map[string]int64) {
			w(f, 2)
			wg.Done()
		}(f)
	}
	wg.Wait()
	//t.Errorf("%d", len(s.Containers.Containers))
}


func TestClose(t *testing.T) {
	s.Close()
}

func TestDrop(t *testing.T) {
	err := s.Drop()
	if err != nil {
		t.Errorf("Storage.Drop() has error: %v", err)
	}
}

func addAndAssert(t *testing.T, name string, size int64) {
	rnd := randReader(size)
	f := s.Add(name, rnd, size)
	if f == nil {
		t.Errorf("Storage.Add(%s, %d) return nil", name, size)
	}
	fg, ok := s.Get(name)
	if ! ok {
		t.Errorf("Storage.Get(%s) return false", name)
	}
	if fg != f {
		t.Error("Files not match")
	}
	if fg.Size != size {
		t.Errorf("File sizes mismatch. Actual: %d, expected: %d", size, fg.Size)
	}
	
	// check md5
	h := md5.New()
	io.Copy(h, fg.GetReader())
	
	act := fmt.Sprintf("%x", h.Sum(nil))
	exp := fmt.Sprintf("%x", f.Md5)
	
	if act != exp {
		t.Error("Md5 file %s mismatch. Actual: %s, expected: %s", name, act, exp)
	}
}


func randReader(n int64) io.Reader {
	return io.LimitReader(rand.Reader, n)
}

func createRandFiles(count int) {
	randFiles = make(map[string]int64, count)
	for len(randFiles) < count {
		randFiles[fmt.Sprintf("rf-%d", mrand.Int())] = int64(mrand.Intn(5000) + 10)  
	}
}


func storageConf() *config.Config {
	mb, _ := utils.BytesFromString("1m")
	return &config.Config{
		// Data path
		DataPath : "",
		ContainerSize : 3 * mb,
		MinEmptySpace : 1  * mb,
	}
}

