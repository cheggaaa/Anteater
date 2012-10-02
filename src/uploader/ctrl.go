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
	Response []*File `json:"response,omitempty"`
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
	return http.PostForm("status", url.Values{"token":{token},"data":{jsonData}})
}
