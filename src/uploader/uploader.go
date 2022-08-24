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
	"github.com/cheggaaa/Anteater/src/aelog"
	"github.com/cheggaaa/Anteater/src/config"
	"github.com/cheggaaa/Anteater/src/storage"
	"io"
	"net/http"
)

type Uploader struct {
	conf *config.Config
	stor *storage.Storage
	ctrl *Ctrl
}

func NewUploader(c *config.Config, s *storage.Storage) *Uploader {
	return &Uploader{conf: c, stor: s, ctrl: NewCtrl(c.UploaderCtrlUrl)}
}

func (u *Uploader) TryRequest(r *http.Request, w http.ResponseWriter) (status bool, err error, errCode int) {
	if !u.conf.UploaderEnable {
		return
	}
	token := r.URL.Query().Get(u.conf.UploaderParamName)
	if token == "" {
		return
	}

	aelog.Debugf("Uploader: token found: %s\n", token)

	status = true

	write := func(token string, params interface{}, err error) {
		var resp *http.Response
		if err != nil {
			resp, err = u.ctrl.SetStatusErr(token, err)
		} else {
			resp, err = u.ctrl.SetStatus(token, params)
		}
		if err != nil {
			errCode = 500
			return
		}
		w.Header().Set("Content-Type", "application/x-javascript")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		err = nil
	}

	files, err := u.ctrl.GetParams(token)
	if err != nil {
		if err == ErrTokenNotFound {
			aelog.Debugln("Uploader: ctrl return not found")
			errCode = 404
		} else {
			errCode = 500
			aelog.Debugf("Uploader: ctrl return error: %v\n", err)
			write(token, nil, err)
		}
		return
	}

	err = files.Upload(u.conf, u.stor, r, w)

	if err != nil {
		aelog.Debugf("Uploader: files.Upload has err: %v\n", err)
		errCode = 500
		write(token, nil, err)
		return
	}
	write(token, files, nil)
	aelog.Debugln("Uploader: files.Upload done")
	return
}