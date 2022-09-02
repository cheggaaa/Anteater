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

package module

import (
	"archive/zip"
	"fmt"
	"github.com/cheggaaa/Anteater/aelog"
	"github.com/cheggaaa/Anteater/storage"
	"github.com/cheggaaa/Anteater/utils"
	"net/http"
	"strings"
)

const (
	UNZIP_NO     = 0
	UNZIP_NORMAL = 1
	UNZIP_FORCE  = 2
)

type unZip struct{}

func (u unZip) OnSave(file *storage.File, w http.ResponseWriter, r *http.Request, s *storage.Storage) (e error) {
	mode := u.needUnZip(r)
	if mode == UNZIP_NO {
		return
	}
	aelog.Debugf("UnZip: start unzip(%d) files to: %s (%s)", mode, file.Name, utils.HumanBytes(file.FSize))
	if mode == UNZIP_FORCE {
		s.DeleteChilds(file.Name)
	}

	filesCount, filesSize, err := u.unZipTo(file.Name, file, s)
	if err != nil {
		aelog.Debugf("UnZip: error: %v", err)
		w.Header().Add("X-Ae-Unzip-Error", err.Error())
		return
	}
	aelog.Debugf("UnZip: unziped: %s (%d files, %s files size)", file.Name, filesCount, utils.HumanBytes(filesSize))
	w.Header().Add("X-Ae-Unzip-Count", fmt.Sprint(filesCount))
	w.Header().Add("X-Ae-Unzip-Size", utils.HumanBytes(filesSize))
	return
}

func (u unZip) OnCommand(command, filename string, w http.ResponseWriter, r *http.Request, s *storage.Storage) (cont bool, e error) {
	if command != "unzip" {
		return true, nil
	}
	if file, ok := s.Get(filename); ok {
		return true, u.OnSave(file, w, r, s)
	}
	return true, nil
}

func (u unZip) unZipTo(to string, f *storage.File, s *storage.Storage) (filesCount, filesSize int64, err error) {

	if err = f.Open(); err != nil {
		return
	}
	defer f.Close()

	z, err := zip.NewReader(f.GetReader(), f.FSize)
	if err != nil {
		return
	}

	for _, zf := range z.File {
		if !zf.FileInfo().IsDir() {
			fs, e := u.saveFile(to, zf, s)
			if e != nil {
				err = e
				return
			}
			filesSize += fs
			filesCount++
		}
	}

	return
}

func (u unZip) saveFile(to string, zf *zip.File, s *storage.Storage) (fs int64, err error) {
	if zf.FileInfo().Size() == 0 {
		return
	}
	fname := strings.Trim(to, "/") + "/" + strings.Trim(zf.Name, "/")
	reader, err := zf.Open()
	if err != nil {
		return
	}
	defer reader.Close()
	if f, err := s.Add(fname, reader, zf.FileInfo().Size()); err == nil {
		return f.FSize, nil
	}
	return
}

func (u unZip) needUnZip(r *http.Request) int {
	switch strings.ToLower(r.Header.Get("X-Ae-Unzip")) {
	case "1", "true", "yes":
		return UNZIP_NORMAL
	case "force":
		return UNZIP_FORCE
	}
	return UNZIP_NO
}
