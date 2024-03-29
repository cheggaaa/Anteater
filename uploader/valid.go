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
	"github.com/cheggaaa/Anteater/temp"
)

var ValidErrTooLarge = errors.New("File too large")

type Valid struct {
	MaxSize   int64    `json:"max_size,omitempty"`
	MimeTypes []string `json:"mime_types,omitempty"`
}

func (v *Valid) HasError(tf *temp.File) (err error) {
	if v.MaxSize > 0 {
		if tf.Size > v.MaxSize {
			err = ValidErrTooLarge
		}
	}
	if len(v.MimeTypes) > 0 {
		found := false
		for _, mt := range v.MimeTypes {
			if mt == tf.MimeType {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("Invalid file type: %s", tf.MimeType)
		}
	}

	return
}