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

package anteater

import (
	"net/http"
	"strings"
	"net/url"
	"encoding/json"
	"io"
	"io/ioutil"
	"errors"
	"mime/multipart"
	"os"
	"fmt"
)

type UploaderParams struct {
	Response UploaderFiles `json:"response,omitempty"`
}

type UploaderValid struct {
	MaxSize    int64    `json:",max_size,omitempty"`
	MimeTypes  []string `json:"mime_types,omitempty"`
	UserAgent  string   `json:"user_agent,omitempty"`
}

type UploaderFiles []*UploaderFile

type UploaderFile struct {
	// name for save to anteater
	Name	string `json:"name,omitempty"`
	// type - image or file
	Type    string `json:"type,omitempty"`
	// field name in form
	Field   string `json:"field,omitempty"`
	// validate
	Valid   *UploaderValid `json:"valid,omitempty"`
	// file state
	State   *UploaderFileState `json:"state,omitempty"`
		
	// Only for images
	// GIF, JPG, PNG
	Format  string `json:"format,omitempty"`
	// image width
	Width   int `json:"width,omitempty"`
	// image height
	Height  int `json:"height,omitempty"`
	// image quality (for jpg)
	Quality int `json:"quality,omitempty"`
	// need crop
	Crop    bool `json:"crop,omitempty"`
	
	// Protected
	_is_uploaded bool
	_fi          *FileInfo
	_tf          *os.File
	_size        int64
}

type UploaderStatusOk struct {
	Response []*UploaderFile `json:"response,omitempty"`
}

type UploaderStatusError struct {
	Error *UploaderError `json:"error,omitempty"`
}

type UploaderFileState struct {
	Uploaded bool  `json:"uploaded,omitempty"`
	Size     int64 `json:"size,omitempty"`
	Md5      string `json:"md5,omitempty"`
}

type UploaderError struct {
	Code int `json:"code,omitempty"`
	Msg string `json:"message,omitempty"`
}

func UploaderHandle(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get(Conf.UploaderTokenName)
	
	if token == "" {
		HttpError(w, 404)
		return
	}
	
	write := func(token string, params interface{}) {
		resp, err := UploaderSetStatus(token, params)
		if err != nil {
			Log.Warnln(err)
			HttpError(w, 500)
			return
		}
		w.Header().Set("Content-Type", "application/x-javascript")
		w.Header().Set("Server", SERVER_SIGN)
		io.Copy(w, resp.Body)
		w.WriteHeader(resp.StatusCode)
	}
	
	Log.Debugln("Start upload, token:", token);
	
	params, err := UploaderGetParams(token)
	if err != nil {
		Log.Debugln(err)
		write(token, &UploaderStatusError{&UploaderError{500, err.Error()}})
		return
	}

	err = params.Upload(r)
	
	if err != nil {
		Log.Warnln(err)
		write(token, &UploaderStatusError{&UploaderError{500, err.Error()}})
		return
	}
	
	write(token, params.Response)
	return
}


func UploaderGetParams(token string) (*UploaderParams, error) {
	r, err := UploaderRequest("params", &url.Values{"token" : {token}})
	if err != nil {
		return nil, err
	}
	var params UploaderParams
	switch r.StatusCode {
		case 200:
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			errj := json.Unmarshal(buf, &params)
			if errj != nil {
				return nil, errj
			}
		default:
			return nil, errors.New(r.Status)
	}
	return &params, nil
}

func UploaderSetStatus(token string, params interface{}) (*http.Response, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	jsonData := string(data)
	return UploaderRequest("status", &url.Values{"token":{token},"data":{jsonData}})
}

func UploaderRequest(uri string, params *url.Values) (*http.Response, error) {
	req_url := strings.TrimRight(Conf.UploaderCtrlUrl, "/") + "/" + strings.TrimLeft(uri, "/") + ".json"
	return http.PostForm(req_url, *params)
}

func UploaderLoadFormFile(f multipart.File) (tf *os.File, size int64, err error) {
	tf, err = ioutil.TempFile(os.TempDir(), "anteater-")
	if err != nil {
		return
	}
	size, err = io.Copy(tf, f)
	return
}

// Uploader params
func (up *UploaderParams) Upload(r *http.Request) error {
	var err error
	var err_n int
	for i, uf := range up.Response {
		err = uf.Upload(r, &up.Response)
		if err != nil {
			err_n = i
			break;
		}
	}
	if err != nil {
		for i := 0; i < err_n; i++ {
			up.Response[i].Delete()
		}
		return err
	}
	for _, uf := range up.Response {
		uf.Clean()
	}
	return nil
}


// uploader file
func (uf *UploaderFile) Upload(r *http.Request, ufs *UploaderFiles) error {
	Log.Debugln("Load form file:", uf.Field)
	mf, _, err := r.FormFile(uf.Field)
	if err != nil {
		return err
	}
	
	uf._tf, uf._size, err = UploaderLoadFormFile(mf)
	if err != nil {
		return err
	}
	uf._tf.Seek(0, 0)
	
	switch(strings.ToLower(uf.Type)) {
		case "file":
			Log.Debugln("Type: file")
			break
		case "image":
			Log.Debugln("Type: image")
			i, err := ImageIdenty(uf._tf.Name())
			if err != nil {
				return err
			}
			uf._tf.Close()
			if uf.Crop {
				i.Crop(uf.Format, uf.Width, uf.Height, uf.Quality)
			} else {
				i.Resize(uf.Format, uf.Width, uf.Height, uf.Quality)
			}
			Log.Debugln("open: ", i.Filename)
			uf._tf, err = os.Open(i.Filename)
			defer os.Remove(i.Filename)
			if err != nil {
				return err
			}
			defer uf._tf.Close()
			finfo, err := uf._tf.Stat()
			if err != nil {
				return err
			}
			uf._size = finfo.Size()
			break
	}
	
	uf._fi, err = WriteFileToStorage(uf._tf, uf.Name, uf._size)
	if err != nil {
		return err
	}
	uf._is_uploaded = true
	uf.State = &UploaderFileState{true, uf._size, fmt.Sprintf("%x", uf._fi.Md5)}
	return nil
}

func (uf *UploaderFile) Delete() {
	HttpDeleteFile(uf.Name, nil)		
	if uf.State != nil {
		uf.State.Uploaded = false
	}
	return 
}

func (uf *UploaderFile) Prepare(ufs *UploaderFiles) error {
	return nil
}

func (uf *UploaderFile) IsValid() bool {
	return true
}

func (uf *UploaderFile) Clean() {
	if uf._tf != nil {
		i, err := uf._tf.Stat()
		if err == nil {
			Log.Debugln(i.Name())
			os.Remove(i.Name())
		}
	}
}