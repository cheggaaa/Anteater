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
	"sync"
)

var ctypes = make(map[string]*CType)
var ctypeM = &sync.Mutex{}

type CType string

func (ct *CType) String() string {
	return string(*ct)
}

func getCtype(ctype string) *CType {
	ctypeM.Lock()
	defer ctypeM.Unlock()
	if ct, ok := ctypes[ctype]; ok {
		return ct
	}
	ct := CType(ctype)
	ctypes[ctype] = &ct
	return ctypes[ctype]
}
