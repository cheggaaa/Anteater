package temp

import (
	"os"
	"io"
	"io/ioutil"
	"cnst"
	"mime/multipart"
	"errors"
)

var TmpPrefix = "ae-" + cnst.VERSION + "-"

type File struct {
	File *os.File
	Filename string
	Size int64
	tmpdir string
}


func NewFile(tmpdir string) *File {
	if tmpdir == "" {
		tmpdir = os.TempDir()
	}
	return &File{tmpdir:tmpdir}
}


func (f *File) LoadFromForm(ff multipart.File) (err error) {
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

func (f *File) Close() (err error) {
	if f.Filename == "" { 
		err = f.Close()
		err = os.Remove(f.Filename)
		f.Filename = ""
		f.File = nil
	}
	return
}

func (f *File) setState() (err error) {
	if f.File != nil {
		i, e := f.File.Stat()
		if e != nil {
			err = e
			return
		}
		f.Filename = i.Name()
		f.Size = i.Size()
		f.File.Seek(0, 0)
	}
	return
}