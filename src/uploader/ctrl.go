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
	"strings"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil"
	"errors"
	"fmt"
)

var ErrTokenNotFound = errors.New("Ctrl server return not found")


type Ctrl struct {
	url string
}

type StatusOk struct {
	Response *Files `json:"response,omitempty"`
}

type StatusError struct {
	Error *Error `json:"error,omitempty"`
}

type Params struct {
	Response *Files `json:"response,omitempty"`
}

func NewCtrl(url string) *Ctrl {
	url = strings.TrimRight(url, "/") 
	return &Ctrl{url}
}


func (c *Ctrl) GetParams(token string) (files *Files, err error) {
	curl := c.url + "/params.json"
	r, err := http.PostForm(curl, url.Values{"token":{token}})
	if err != nil {
		return 
	}
	defer r.Body.Close()
	
	if r.StatusCode != 200 {
		if r.StatusCode == 404 {
			err = ErrTokenNotFound
		} else {
			err = errors.New(fmt.Sprintf("Ctrl server return non-200 status: %s", r.Status))
		}
		return
	}
	
	params := new(Params)
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return
	}
	
	files = params.Response	
	return
}

func (c *Ctrl) SetStatus(token string, params interface{}) (*http.Response, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	jsonData := string(data)
	curl := c.url + "/status.json"
	return http.PostForm(curl, url.Values{"token":{token},"data":{jsonData}})
}

func (c *Ctrl) SetStatusErr(token string, err error) (*http.Response, error) {
	statusErr := &StatusError{&Error{Code : 500, Msg : err.Error()}}
	return c.SetStatus(token, statusErr)
}
