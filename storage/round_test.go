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
	"github.com/cheggaaa/Anteater/utils"
	"testing"
)

func TestRound(t *testing.T) {
	c := int32(1000)
	ls := int64(0)
	for i := int32(1); i <= c; i++ {
		size := R.Size(i)
		indx := R.Index(size)
		t.Logf("I%d\t%s", indx, utils.HumanBytes(size))
		if indx != i {
			t.Errorf("Size-index conversion failed! %d vs %d", indx, i)
		}
		if size <= ls {
			t.Errorf("Last size more then actual: %d vs %d", ls, size)
		}
		ls = size
	}
}
