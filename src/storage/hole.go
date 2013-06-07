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

type Hole struct {
	// Next and Prev
	prev, next Space
	// Offset
	Off int64
	// Index
	Indx int
}

// implement Space
func (h *Hole) Next() Space {
	return h.next
}

func (h *Hole) Prev() Space {
	return h.prev
}

func (h *Hole) SetNext(s Space) {
	h.next = s
}

func (h *Hole) SetPrev(s Space) {
	h.prev = s
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

func newHole(offset, size int64) *Hole {
	return &Hole{Off: offset, Indx: R.Index(size)}
}
