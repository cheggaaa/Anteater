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
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

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
	MarshalTo(wr io.Writer) error
}

func UnmarshallSpace(rd *bufio.Reader) (s Space, err error) {
	var f *File
	tp, err := rd.ReadByte()
	if err != nil {
		return
	}
	if tp == 1 {
		f = &File{}
		nl, e := binary.ReadUvarint(rd)
		if e != nil {
			return nil, e
		}
		tm, e := binary.ReadUvarint(rd)
		if e != nil {
			return nil, e
		}
		sz, e := binary.ReadUvarint(rd)
		if e != nil {
			return nil, e
		}
		var name = make([]byte, nl)
		if _, err = io.ReadFull(rd, name); err != nil {
			return
		}
		f.Md5 = make([]byte, 16)
		if _, err = io.ReadFull(rd, f.Md5); err != nil {
			return
		}
		f.Name = string(name)
		f.FSize = int64(sz)
		f.Time = time.Unix(int64(tm), 0)
		tp, e = rd.ReadByte()
		if e != nil {
			return nil, e
		}
	}

	if tp != 0 {
		return nil, fmt.Errorf("unexpected hole type: %v; %+v", tp, f)
	}
	indx, err := binary.ReadUvarint(rd)
	if err != nil {
		return
	}
	offset, err := binary.ReadUvarint(rd)
	if err != nil {
		return
	}
	h := Hole{Indx: int32(indx), Off: int64(offset)}
	if f != nil {
		f.Hole = h
		return f, nil
	}
	return &h, nil
}
