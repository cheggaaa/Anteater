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

package config

import (
	"github.com/kless/goconfig/config"
	"strings"
	"errors"
	"utils"
	"cnst"
)

type Config struct {
	// Data path
	DataPath      string
	ContainerSize int64
	MinEmptySpace  int64
	
	// Http
	HttpWriteAddr string
	HttpReadAddr  string
	ETagSupport   bool
	ContentRange  int64
	StatusJson    string
	StatusHtml    string
	
	// Rpc
	RpcAddr       string
	
	// Http Headers
	Headers       map[string]string
	
	// Mime Types
	MimeTypes     map[string]string
	
	// Log
	LogLevel      int
	LogFile		  string
	
	// Uploader
	UploaderEnable    bool
	UploaderCtrlUrl   string
	UploaderTokenName string
}


func (conf *Config) ReadFile(filename string) (err error) {
	c, err := config.ReadDefault(filename)
	if err != nil {
		return
	}
	
	// Data path
	conf.DataPath, err = c.String("data", "path")
	if err != nil {
		return
	}
	conf.DataPath = strings.TrimRight(conf.DataPath, "/") + "/"
	
	// Container size
	s, err := c.String("data", "container_size")
	if err != nil {
		return
	}
	conf.ContainerSize, err = utils.BytesFromString(s)
	
	// Min empty space
	s, err = c.String("data", "min_empty_space")
	if err != nil {
		return
	}
	conf.MinEmptySpace, err = utils.BytesFromString(s)
	
	// Http write addr
	conf.HttpWriteAddr, err = c.String("http", "write_addr")
	if err != nil {
		return
	}
	
	// Http read addr
	conf.HttpReadAddr, err = c.String("http", "read_addr")
	if err != nil || len(conf.HttpReadAddr) == 0 {
		conf.HttpReadAddr = conf.HttpWriteAddr
	}
	
	// ETag flag
	conf.ETagSupport, err = c.Bool("http", "etag")
	if err != nil {
		conf.ETagSupport = true
	}
	
	// Range support
	cr, err := c.String("http", "content_range")
	if err != nil {
		cr = "5M"
	}
	conf.ContentRange, err = utils.BytesFromString(cr)
	
	// Status json
	conf.StatusJson, err = c.String("http", "status_json")
	if err != nil {
		conf.StatusJson = ""
	}
	
	// Status html
	conf.StatusHtml, err = c.String("http", "status_html")
	if err != nil {
		conf.StatusHtml = ""
	}
	
	conf.RpcAddr, err = c.String("rpc", "addr")
	if err != nil {
		conf.RpcAddr = ":32032"
	}

	
	// Headers	
	headers := make(map[string]string, 0)
	hOpts, err := c.Options("http-headers")
	if err == nil {
		for _, o := range(hOpts) {
			v, err := c.String("http-headers", o)
			if err == nil && len(v) > 0 {
				headers[o] = v
			} 
		}
	}
	if _, ok := headers["Server"]; !ok {
		headers["Server"] = cnst.SIGN
	}
	
	conf.Headers = headers
	
	// Mime	
	mimeTypes := make(map[string]string, 0)
	mOpts, err := c.Options("mime-types")
	if err == nil {
		for _, o := range(mOpts) {
			v, err := c.String("mime-types", o)
			if err == nil && len(v) > 0 {
				mimeTypes["." + o] = v
			} 
		}
	}
	
	conf.MimeTypes = mimeTypes
	
	// Log level
	levels := map[string]int {
		"debug" : cnst.LOG_DEBUG,
		"info"  : cnst.LOG_INFO,
		"warn"  : cnst.LOG_WARN,
	}
	llv, err := c.String("log", "level")
	if err != nil {
		llv = "info"
	}
	logLevel, ok := levels[llv]
	if ! ok {
		logLevel = levels["info"]
	}	
	conf.LogLevel = logLevel
	
	// Log file
	conf.LogFile, err = c.String("log", "file")
	if err != nil {
		conf.LogFile = ""
	}
	
	// Uploader
	conf.UploaderEnable, err = c.Bool("uploader", "enable")
	if err != nil {
		conf.UploaderEnable = false
	}
	
	conf.UploaderCtrlUrl, err = c.String("uploader", "ctrl_url")
	if err != nil && conf.UploaderEnable {
		err = errors.New("Incorrect uploader ctrl_url:" + err.Error())
		return
	}
	
	conf.UploaderTokenName, err = c.String("uploader", "token_name")
	if err != nil && conf.UploaderEnable {
		err = errors.New("Incorrect uploader token_name:" + err.Error())
		return
	}
	
	err = nil
	return
}
