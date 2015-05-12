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
	"config"
	"errors"
	"net/http"
	"storage"
	"temp"
)

type Files []*File

type TmpFiles struct {
	r      *http.Request
	fields map[string]*temp.File
	result map[string]*temp.File
	tmpDir string
}

func (fs *Files) Upload(conf *config.Config, stor *storage.Storage, r *http.Request, w http.ResponseWriter) (err error) {
	// make temp files object
	tmpfs := &TmpFiles{r: r, fields: make(map[string]*temp.File), result: make(map[string]*temp.File), tmpDir: conf.TmpDir}
	defer tmpfs.Close()

	// upload
	for _, f := range *fs {
		err = f.Upload(tmpfs)
		if err != nil {
			return
		}
	}

	result := tmpfs.Result()

	// add to storage and set stae
	for name, tf := range result {
		file, _ := stor.Add(name, tf.File, tf.Size)
		for _, f := range *fs {
			if f.Name == name {
				f.SetState(file)
			}
		}
	}

	return
}

func (t *TmpFiles) GetByField(field string) (f *temp.File, err error) {
	f = t.fields[field]
	if f == nil {
		url := t.r.PostFormValue(field + "_url")
		mf, _, e := t.r.FormFile(field)
		if e != nil {
			err = e
			return
		}
		if mf == nil {
			err = errors.New("File " + field + " not found")
			return
		}
		f = temp.NewFile(t.tmpDir)
		if url != "" {
			if err = f.LoadFromUrl(url); err != nil {
				return
			}
		} else {
			if err = f.LoadFromForm(mf); err != nil {
				return
			}
		}
		t.fields[field] = f
	}
	return
}

func (t *TmpFiles) SetResult(name string, file *temp.File) {
	t.result[name] = file
}

func (t *TmpFiles) Result() map[string]*temp.File {
	return t.result
}

func (t *TmpFiles) Close() {
	for _, f := range t.fields {
		f.Close()
	}
	for _, f := range t.result {
		f.Close()
	}
}
