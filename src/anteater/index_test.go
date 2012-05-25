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

package anteater

import (
	"testing"
)

var (
	info1 = &FileInfo{Id: 1}
	info2 = &FileInfo{Id: 2}
	info3 = &FileInfo{Id: 3}
)

func TestIndexSetGet(t *testing.T) {
	for n, _ := range Index {
		delete(Index, n)
	}

	IndexSet("1", info1)
	in1, ok1 := IndexGet("1")
	if ok1 != true {
		t.Errorf("ok1 != true\n")
	}	
	if in1.Id != int64(1) {
		t.Errorf("in1.Id == 1, result:%d\n", in1.Id)
	}	
	if len(Index) != 1 {
		t.Errorf("len(Index) must be 1, result:%d\n", len(Index))
	}
	
	IndexSet("1", info2)
	in2, ok2 := IndexGet("1")
	if ok2 != true {
		t.Errorf("ok2 != true\n")
	}	
	if in2.Id != int64(2) {
		t.Errorf("in2.Id == 2, result:%d\n", in2.Id)
	}	
	if len(Index) != 1 {
		t.Errorf("len(Index) must be 1, result:%d\n", len(Index))
	}
	
	IndexSet("3", info3)
	in3, ok3 := IndexGet("3")
	if ok3 != true {
		t.Errorf("ok3 != true\n")
	}	
	if in3.Id != int64(3) {
		t.Errorf("in3.Id == 3, result:%d\n", in3.Id)
	}	
	if len(Index) != 2 {
		t.Errorf("len(Index) must be 2, result:%d\n", len(Index))
	}
}

func TestIndexDelete(t *testing.T) {
	IndexSet("1", info1)
	IndexSet("2", info2)
	IndexSet("3", info3)
	if len(Index) != 3 {
		t.Errorf("len(Index) must be 3, result:%d\n", len(Index))
	}
	
	in2, ok2 := IndexDelete("2")
	if ok2 != true {
		t.Errorf("ok2 != true\n")
	}	
	if in2.Id != int64(2) {
		t.Errorf("in2.Id == 2, result:%d\n", in2.Id)
	}

	in2, ok2 = IndexDelete("2")
	if ok2 != false {
		t.Errorf("ok2 != false\n")
	}	
	if in2 != nil {
		t.Errorf("in2 must be nil, result:%v", in2)
	}
	
	if len(Index) != 2 {
		t.Errorf("len(Index) must be 2, result:%d\n", len(Index))
	}
}

