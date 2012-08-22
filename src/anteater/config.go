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
	"github.com/kless/goconfig/config"
	"errors"
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


func LoadConfig(filename string) (*Config, error) {
	c, err := config.ReadDefault(filename)
	if err != nil {
		return nil, err
	}
	
	// Data path
	dataPath, err := c.String("data", "path")
	if err != nil {
		return nil, err
	}
	if len(dataPath) == 0 {
		return nil, errors.New("Empty data path in config " + filename)
	}
	
	// Container size
	s, err := c.String("data", "container_size")
	if err != nil {
		return nil, err
	}
	containerSize, err := GetSizeFromString(s)
	
	// Min empty space
	s, err = c.String("data", "min_empty_space")
	if err != nil {
		return nil, err
	}
	minEmptySpace, err := GetSizeFromString(s)
	
	// Http write addr
	httpWriteAddr, err := c.String("http", "write_addr")
	if err != nil {
		return nil, err
	}
	
	// Http read addr
	httpReadAddr, err := c.String("http", "read_addr")
	if err != nil || len(httpReadAddr) == 0 {
		httpReadAddr = httpWriteAddr
	}
	
	// ETag flag
	etagSupport, err := c.Bool("http", "etag")
	if err != nil {
		etagSupport = true
	}
	
	// Range support
	cr, err := c.String("http", "content_range")
	if err != nil {
		cr = "5M"
	}
	contentRange, err := GetSizeFromString(cr)
	
	// Status json
	statusJson, err := c.String("http", "status_json")
	if err != nil {
		statusJson = ""
	}
	
	// Status html
	statusHtml, err := c.String("http", "status_html")
	if err != nil {
		statusHtml = ""
	}
	
	rpcAddr, err := c.String("rpc", "addr")
	if err != nil {
		rpcAddr = ":32032"
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
		headers["Server"] = SERVER_SIGN
	}
	
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
	
	// Log level
	levels := map[string]int {
		"debug" : LOG_DEBUG,
		"info"  : LOG_INFO,
		"warn"  : LOG_WARN,
	}
	llv, err := c.String("log", "level")
	if err != nil {
		llv = "info"
	}
	logLevel, ok := levels[llv]
	if ! ok {
		logLevel = levels["info"]
	}
	
	// Log file
	logFile, err := c.String("log", "file")
	if err != nil {
		logFile = ""
	}
	
	// Uploader
	uploaderEnable, err := c.Bool("uploader", "enable")
	if err != nil {
		uploaderEnable = false
	}
	
	uploaderCtrlUrl, err := c.String("uploader", "ctrl_url")
	if err != nil && uploaderEnable {
		return nil, errors.New("Incorrect uploader ctrl_url:" + err.Error())
	}
	
	uploaderTokenName, err := c.String("uploader", "token_name")
	if err != nil && uploaderEnable {
		return nil, errors.New("Incorrect uploader token_name:" + err.Error())
	}
	
	return &Config{
		dataPath,
		containerSize,
		minEmptySpace,
		httpWriteAddr,
		httpReadAddr,
		etagSupport,
		contentRange,
		statusJson,
		statusHtml,
		rpcAddr,
		headers,
		mimeTypes,
		logLevel,
		logFile,
		uploaderEnable,
		uploaderCtrlUrl,
		uploaderTokenName,
	}, nil
}
