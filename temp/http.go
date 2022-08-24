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
	"net/http"
	"io"
	"errors"
	"fmt"
)

func (f *File) LoadFromUrl(url string) (err error) {
	r, err := http.Get(url)
	if err != nil {
		return
	}
	defer r.Body.Close()
	
	err = f.Create()
	if err != nil {
		return
	}
	
	if r.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("server return non-200 status: %s", r.Status))
		return
	}
	_, err = io.Copy(f.File, r.Body)
	
	if err == nil {
		err = f.setState()
	}
	
	return
}