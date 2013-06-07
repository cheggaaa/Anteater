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
	"aelog"
	"config"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"testing"
	"utils"
	"time"
)

var S *Storage
var testConfig = &config.Config{
	Md5Header:     true,
	ContainerSize: 1024 * 1024 * 10, // 10 Mb
	DataPath:      os.TempDir() + "/",
}

func init() {
	mrand.Seed(time.Now().UnixNano())
}

func TestOpen(t *testing.T) {
	aelog.InitDefault(aelog.LOG_DEBUG)

	S = new(Storage)
	S.Init(testConfig)
	if err := S.Open(); err != nil {
		t.Errorf("Can't open storage: %v", err)
	}
}

func TestFirstAdd(t *testing.T) {
	addAndAssert(t, "first", "1m")
	if len(S.Containers) != 1 {
		t.Errorf("Unexpected containers count: %d", len(S.Containers))
	}

	t.Log(1)

	var C *Container
	for _, C = range S.Containers {
	}

	if C.FileCount != 1 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}

	if ok := S.Delete("first"); !ok {
		t.Errorf("Can't delete file")
	}

	if C.FileCount != 0 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}
}

func TestAppend(t *testing.T) {
	c := 10
	for i := 0; i < c; i++ {
		addAndAssert(t, fmt.Sprint(i), "1m")
	}

	if len(S.Containers) != 1 {
		t.Errorf("Unexpected containers count: %d", len(S.Containers))
	}

	var C *Container
	for _, C = range S.Containers {
	}

	if C.FileCount != int64(c) {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}
}

func TestDeleteLast(t *testing.T) {
	if ok := S.Delete("9"); !ok {
		t.Errorf("Can't delete file")
	}
	var C *Container
	for _, C = range S.Containers {
	}

	if C.FileCount != 9 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}

	addAndAssert(t, "10", "1m")
	if C.FileCount != 10 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}

	if ok := S.Delete("8"); !ok {
		t.Errorf("Can't delete file")
	}

	if C.FileCount != 9 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}

	addAndAssert(t, "8", "1m")
	if C.FileCount != 10 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}
}

func TestAllocInsert(t *testing.T) {
	if ok := S.Delete("8"); !ok {
		t.Errorf("Can't delete file")
	}
	var C *Container
	for _, C = range S.Containers {
	}

	C.Print()

	if C.FileCount != 9 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}

	for i := 0; i < 10; i++ {
		addAndAssert(t, fmt.Sprintf("s%d", i), "90k")
	}

	if C.FileCount != 19 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}
}

func TestRemoveAll(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Logf("Delete: s%d", i)
		if ok := S.Delete(fmt.Sprintf("s%d", i)); !ok {
			t.Errorf("Can't delete file: s%d", i)
		}
		if i != 8 {
			if i == 9 {
				i = 10
			}
			t.Logf("Delete: %d", i)
			if ok := S.Delete(fmt.Sprint(i)); !ok {
				t.Errorf("Can't delete file: %d", i)
			}
		}

	}

	var C *Container
	for _, C = range S.Containers {
	}

	C.Print()

	if C.FileCount != 0 {
		t.Errorf("Unexpected files count: %d", C.FileCount)
	}
	if C.last != nil {
		t.Errorf("Last must be nil")
	}
	if C.holeIndex.Count != 0 {
		t.Errorf("Unexpected holes count: %d", C.holeIndex.Count)
	}
}

func TestCreateContainer(t *testing.T) {
	for i := 0; i < 160; i++ {
		addAndAssert(t, fmt.Sprintf("%d", i), "128k")
	}
	if len(S.Containers) != 2 {
		t.Errorf("Must be 2 containers, expected: %d", len(S.Containers))
	}

	for _, c := range S.Containers {
		if err := c.Check(); err != nil {
			t.Errorf("Container check failed: %v", err)
		}
	}
	for i := 0; i < 160; i++ {
		S.Delete(fmt.Sprintf("%d", i))
	}
}

func TestDumpRestore(t *testing.T) {
	var C *Container
	for _, C = range S.Containers {
	}

	fc := C.FileCount
	hc := C.holeIndex.Count
	ic := S.Index.Count()

	S.Dump()
	S.Close()
	S = new(Storage)
	S.Init(testConfig)
	if err := S.Open(); err != nil {
		t.Errorf("Can't open storage: %v", err)
	}
	for _, C = range S.Containers {
	}

	C.Print()

	if err := C.Check(); err != nil {
		t.Errorf("Container check failed: %v", err)
	}
	if fc != C.FileCount {
		t.Errorf("Files count mismatched: %d vs %d", fc, C.FileCount)
	}
	if hc != C.holeIndex.Count {
		t.Errorf("Holes count mismatched: %d vs %d", hc, C.holeIndex.Count)
	}
	if ic != S.Index.Count() {
		t.Errorf("Index count mismatched: %d vs %d", ic, S.Index.Count())
	}
}

func TestRandomUpdate(t *testing.T) {
	count := 1000
	for i := 0; i < count; i++ {
		name := fmt.Sprint(mrand.Intn(100))
		size := fmt.Sprint(mrand.Intn(mrand.Intn(mrand.Intn(1024 * 100)+1)+1)+1)
		S.Delete(name)
		addAndAssert(t, name, size)
	}
	if err := S.Check(); err != nil {
		t.Error(err)
	}
}

func TestDrop(t *testing.T) {
	S.Drop()
}

func addAndAssert(t *testing.T, name string, sizeS string) {
	size, _ := utils.BytesFromString(sizeS)
	rnd := randReader(size)
	f, err := S.Add(name, rnd, size)
	if err != nil {
		t.Errorf("Storage.Add(%s, %d) return err: %v", name, size, err)
	}
	fg, ok := S.Get(name)
	if !ok {
		t.Errorf("Storage.Get(%s) return false", name)
	}
	fg.Open()
	defer fg.Close()
	if fg != f {
		t.Error("Files not match")
	}
	if fg.FSize != size {
		t.Errorf("File sizes mismatch. Actual: %d, expected: %d", size, fg.FSize)
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
