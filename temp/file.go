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

package temp

import (
	"errors"
	"github.com/cheggaaa/Anteater/cnst"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

var TmpPrefix = "ae-" + cnst.VERSION + "-"

type File struct {
	File     *os.File
	OrigName string
	Filename string
	MimeType string
	Size     int64
	tmpdir   string
}

func NewFile(tmpdir string) *File {
	if tmpdir == "" {
		tmpdir = os.TempDir()
	}
	return &File{tmpdir: tmpdir}
}

func (f *File) LoadFromForm(ff multipart.File) (err error) {
	err = f.Create()
	if err != nil {
		return
	}
	_, err = io.Copy(f.File, ff)
	if err == nil {
		err = f.setState()
	}
	return
}

func (f *File) Create() (err error) {
	f.File, err = ioutil.TempFile(f.tmpdir, TmpPrefix)
	return
}

func (f *File) Reopen() (err error) {
	if f.Filename == "" {
		err = errors.New("Can't reopen empty file")
		return
	}
	f.Close()
	f.File, err = os.Open(f.Filename)
	return
}

func (f *File) Disconnect() {
	if f.File != nil {
		f.File.Close()
	}
}

func (f *File) Connect() (err error) {
	f.File, err = os.Open(f.Filename)
	if err == nil {
		err = f.setState()
	}
	return
}

func (f *File) Close() (err error) {
	if f.Filename != "" {
		if f.File != nil {
			err = f.File.Close()
		}
		err = os.Remove(f.Filename)
		f.Filename = ""
		f.File = nil
		f.Size = 0
	}
	return
}

func (f *File) Clone(suffix string) (fl *File, err error) {
	fl = &File{
		Filename: f.Filename + suffix,
	}
	err = fl.setState()
	return
}

func (f *File) setState() (err error) {
	if f.File != nil {
		i, e := f.File.Stat()
		if e != nil {
			err = e
			return
		}
		f.Filename = f.File.Name()
		f.Size = i.Size()
		f.File.Seek(0, 0)
		var buf [512]byte
		n, _ := io.ReadFull(f.File, buf[:])
		b := buf[:n]
		f.MimeType = http.DetectContentType(b)
		f.File.Seek(0, 0)
	}
	return
}
