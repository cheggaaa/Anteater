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

package utils

import (
	"testing"
)

type TestSize struct {
	String string
	Value  int64
}

var TestBytesFromStringSet = []TestSize{
	{"0", 0},
	{"1k", 1024},
	{"3456K", 1024 * 3456},
	{"1m", 1024 * 1024},
	{"23M", 1024 * 1024 * 23},
	{"1g", 1024 * 1024 * 1024},
	{"43G", 43 * 1024 * 1024 * 1024},
	{"1t", 1024 * 1024 * 1024 * 1024},
	{"3t", 3 * 1024 * 1024 * 1024 * 1024},
	{"3T", 3 * 1024 * 1024 * 1024 * 1024},
}

var TestHumanBytesSet = []TestSize{
	{"0 B", 0},
	{"100 B", 100},
	{"1.00 KiB", 1025},
	{"1.01 KiB", 1035},
	{"3.38 MiB", 1024 * 3456},
	{"337.50 MiB", 1024 * 3456 * 100},
	{"3.30 GiB", 1024 * 3456 * 1000},
	{"3.30 TiB", 1024 * 3456 * 1000 * 1024},
}

func TestBytesFromString(t *testing.T) {
	for _, set := range TestBytesFromStringSet {
		res, err := BytesFromString(set.String)
		if err != nil {
			t.Errorf("BytesFromString has error: %v", err)
		}
		if res != set.Value {
			t.Errorf("%s must be == %d, but result: %d", set.String, set.Value, res)
		}
	}
	
	res, err := BytesFromString("ololo")
	if err == nil {
		t.Errorf("BytesFromString must be return error")
	}
	if res != 0 {
		t.Errorf("%s must be == %d, but result: %d", "ololo", 0, res)
	}
}


func TestHumanBytes(t *testing.T) {
	for _, set := range TestHumanBytesSet {
		res := HumanBytes(set.Value)
		if set.String != res {
			t.Errorf("%d must be == %s, but result: %s", set.Value, set.String, res)
		}
	}
}