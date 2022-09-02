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

package uploader

import (
	"errors"
	"fmt"
	"github.com/cheggaaa/Anteater/storage"
	"github.com/cheggaaa/Anteater/temp"
	"github.com/cheggaaa/Anteater/utils"
	"strings"
	"time"
)

type File struct {
	// name for save to anteater
	Name string `json:"name,omitempty"`
	// type - image or file
	Type string `json:"type,omitempty"`
	// field name in form
	Field string `json:"field,omitempty"`
	// validate
	Valid *Valid `json:"valid,omitempty"`
	// file state
	State *FileState `json:"state,omitempty"`

	// Only for images
	// GIF, JPG, PNG
	Format string `json:"format,omitempty"`
	// image width
	Width int `json:"width,omitempty"`
	// image height
	Height int `json:"height,omitempty"`
	// image quality (for jpg)
	Quality int `json:"quality,omitempty"`
	// need crop
	Crop bool `json:"crop,omitempty"`
	// apply optimize for images (png only)
	Optimize bool `json:"optimize,omitempty"`
}

type FileState struct {
	Uploaded bool   `json:"uploaded,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Md5      string `json:"md5,omitempty"`
}

type Error struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"message,omitempty"`
}

func (f *File) Upload(tmpf *TmpFiles) (err error) {
	f.State = &FileState{}
	tf, err := tmpf.GetByField(f.Field)
	if f.checkErr(err) {
		return
	}

	if f.Valid != nil {
		err = f.Valid.HasError(tf)
		if f.checkErr(err) {
			return
		}
	}

	var tfr *temp.File

	switch f.Type {
	case "file":
		tfr, err = f.uploadFile(tf)
		break
	case "image":
		tfr, err = f.uploadImage(tf)
		break
	default:
		err = errors.New("Undefined file type: " + f.Type)
	}

	if f.checkErr(err) {
		return
	}
	f.Name = strings.Replace(f.Name, "%origname%", tfr.OrigName, -1)
	tmpf.SetResult(f.Name, tfr)
	return
}

func (f *File) uploadImage(tf *temp.File) (tfr *temp.File, err error) {
	image, err := utils.Identify(tf.Filename)
	if err != nil {
		return
	}

	tfr, err = tf.Clone("." + fmt.Sprintf("%d%d", time.Now().Unix(), time.Now().UnixNano()))
	if err != nil {
		return
	}
	tfr.Disconnect()
	tf.Disconnect()
	if f.Crop {
		image.Crop(tfr.Filename, f.Format, f.Width, f.Height, f.Quality, f.Optimize)
	} else {
		image.Resize(tfr.Filename, f.Format, f.Width, f.Height, f.Quality, f.Optimize)
	}
	err = tf.Connect()
	if err == nil {
		err = tfr.Connect()
	}
	return
}

func (f *File) uploadFile(tf *temp.File) (tfr *temp.File, err error) {
	tfr = tf
	return
}

func (f *File) SetState(file *storage.File) {
	f.State = &FileState{
		Uploaded: true,
		Size:     file.FSize,
		Md5:      fmt.Sprintf("%x", file.Md5),
	}
}

func (f *File) checkErr(err error) bool {
	if err != nil {
		// store err to state later
		return true
	}
	return false
}
