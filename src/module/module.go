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
	"github.com/cheggaaa/Anteater/src/storage"
	"net/http"
)

type Module interface {
	OnSave(file *storage.File, w http.ResponseWriter, r *http.Request, s *storage.Storage) (err error)
	OnCommand(command, filename string, w http.ResponseWriter, r *http.Request, s *storage.Storage) (cont bool, err error)
}

var modules = make([]Module, 0)

func RegisterModules() {
	modules = append(modules, unZip{}, fileList{})
}

func OnSave(file *storage.File, w http.ResponseWriter, r *http.Request, s *storage.Storage) (err error) {
	for _, m := range modules {
		err = m.OnSave(file, w, r, s)
		if err != nil {
			break
		}
	}
	return
}

func OnCommand(command, filename string, w http.ResponseWriter, r *http.Request, s *storage.Storage) (cont bool, err error) {
	for _, m := range modules {
		cont, err = m.OnCommand(command, filename, w, r, s)
		if !cont || err != nil {
			break
		}
	}
	return
}
