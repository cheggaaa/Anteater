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

package dump

import (
	"fmt"
	"github.com/cheggaaa/Anteater/src/aelog"
	"os"
	"testing"
)

type TD struct {
	Td1 []*TD2
	Td3 *TD3
	Ti  int64
}

type TD2 struct {
	TInt    int
	TString string
	TBool   bool
	Td3     *TD3
}

type TD3 struct {
	Test  bool
	Test2 int64
}

var TestData *TD
var TestCount int = 1000

func TestDump(t *testing.T) {
	aelog.DefaultLogger, _ = aelog.New("", 0)
	for i := 0; i < 5; i++ {
		TestCount -= i * 10
		makeTestData()
		file := "test.dump"
		defer os.Remove(file)
		defer os.Remove(file + ".td")
		n, err := DumpTo(file, TestData)
		if err != nil {
			t.Errorf("Dump has error: %v", err)
		}
		if n <= 0 {
			t.Errorf("Dump write %d bytes. Wrong.", n)
		}

		data := new(TD)
		err, exists := LoadData(file, data)
		if err != nil {
			t.Errorf("LoadData has error: %v", err)
		}
		if !exists {
			t.Errorf("File must be exists")
		}

		if !data.Assert() {
			t.Errorf("Data mismatched")
		}
	}
}

func Benchmark_DumpSave(b *testing.B) {
	makeTestData()
	file := "test.dump"
	defer os.Remove(file)
	defer os.Remove(file + ".td")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, _ := DumpTo(file, TestData)
		b.SetBytes(int64(n))
	}
	b.StopTimer()
}

func Benchmark_DumpRestore(b *testing.B) {
	makeTestData()
	file := "test.dump"
	defer os.Remove(file)
	defer os.Remove(file + ".td")
	n, _ := DumpTo(file, TestData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := new(TD)
		LoadData(file, data)
		b.SetBytes(int64(n))
	}
	b.StopTimer()
}

func makeTestData() {
	td1 := make([]*TD2, TestCount)
	for i := 0; i < TestCount; i++ {
		td1[i] = &TD2{i, fmt.Sprintf("D:%d", i*5), (i % 2) == 0, &TD3{(i % 3) == 0, int64(i * 5)}}
	}
	TestData = &TD{td1, &TD3{}, 12345}
}

func (t *TD) Assert() bool {
	if len(t.Td1) != TestCount {
		return false
	}
	for _, v := range t.Td1 {
		i := v.TInt
		if v.TString != fmt.Sprintf("D:%d", i*5) {
			return false
		}
		if v.TBool != ((i % 2) == 0) {
			return false
		}
		if v.Td3.Test2 != int64(i*5) {
			return false
		}
	}
	return true
}



