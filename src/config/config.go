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
	"time"
	"aelog"
	"mime"
)

type Config struct {
	// Data
	DataPath      string
	ContainerSize int64
	MinEmptySpace  int64
	DumpTime time.Duration
	TmpDir        string
	
	// Http
	HttpWriteAddr string
	HttpReadAddr  string
	HttpWriteTimeout time.Duration
	HttpReadTimeout time.Duration
	ETagSupport   bool
	Md5Header     bool
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
	LogAccess     string
	
	//Downloader
	DownloaderEnable bool
	DownloaderParamName string
	
	
	// Uploader
	UploaderEnable    bool
	UploaderCtrlUrl   string
	UploaderParamName string
	
	// Amazon
	AmazonCFEnable    bool
	AmazonCFDistributionId string
	AmazonCFAuthentication string
	AmazonInvalidationDuration time.Duration
}

// Parse file and set values to config
func (conf *Config) ReadFile(filename string) {
	c, err := config.ReadDefault(filename)
	if err != nil {
		panic(err)
	}
	
	// Data path
	conf.DataPath, err = c.String("data", "path")
	if err != nil {
		panic(err)
	}
	conf.DataPath = strings.TrimRight(conf.DataPath, "/") + "/"
	
	// Container size
	s, err := c.String("data", "container_size")
	if err != nil {
		panic(err)
	}
	conf.ContainerSize, err = utils.BytesFromString(s)
	
	// Min empty space
	s, err = c.String("data", "min_empty_space")
	if err != nil {
		panic(err)
	}
	conf.MinEmptySpace, err = utils.BytesFromString(s)
	
	
	// Min empty space
	s, err = c.String("data", "dump_duration")
	if err != nil {
		conf.DumpTime = time.Minute
	} else {
		conf.DumpTime, err = time.ParseDuration(s)
		if err != nil {
			panic("Incorrect dump time duration")
		}
		if conf.DumpTime <= 0 {
			panic("Incorrect dump time duration")
		} 
	}
	
	// Temp dir
	conf.TmpDir, err = c.String("data", "tmp_dir")
	if err == nil {
		conf.TmpDir = strings.TrimRight(conf.TmpDir, "/")
	}
	
	// Http write addr
	conf.HttpWriteAddr, err = c.String("http", "write_addr")
	if err != nil {
		panic(err)
	}
	
	// Http read addr
	conf.HttpReadAddr, err = c.String("http", "read_addr")
	if err != nil || len(conf.HttpReadAddr) == 0 {
		conf.HttpReadAddr = conf.HttpWriteAddr
	}
	
	// Http write timeout
	s, err = c.String("http", "write_timeout")
	if err == nil {
		conf.HttpWriteTimeout, err = time.ParseDuration(s)
		if err != nil {
			panic("Incorrect http.write_timeout time duration")
		}
		if conf.HttpWriteTimeout <= 0 {
			panic("Incorrect http.write_timeout time duration")
		} 
	}
	
	// Http read timeout
	s, err = c.String("http", "read_timeout")
	if err == nil {
		conf.HttpReadTimeout, err = time.ParseDuration(s)
		if err != nil {
			panic("Incorrect http.read_timeout time duration")
		}
		if conf.HttpReadTimeout <= 0 {
			panic("Incorrect http.read_timeout time duration")
		} 
	}
	
	// ETag flag
	conf.ETagSupport, err = c.Bool("http", "etag")
	if err != nil {
		conf.ETagSupport = true
	}
	
	// Md5 header
	conf.Md5Header, err = c.Bool("http", "md5_header")
	if err != nil {
		conf.Md5Header = true
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
		"debug" : aelog.LOG_DEBUG,
		"info"  : aelog.LOG_INFO,
		"warn"  : aelog.LOG_WARN,
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
	
	// Access log file
	conf.LogAccess, err = c.String("log", "access_log")
	if err != nil {
		conf.LogAccess = ""
	}
	
	// Downloader
	conf.DownloaderEnable, err = c.Bool("downloader", "enable")
	if err != nil {
		conf.DownloaderEnable = false
	}
	if conf.DownloaderEnable {
		conf.DownloaderParamName, err = c.String("downloader", "param_name")
		if err != nil {
			panic("Downloader param_name has error: " + err.Error())
		}
	}
	
	// Uploader
	conf.UploaderEnable, err = c.Bool("uploader", "enable")
	if err != nil {
		conf.UploaderEnable = false
	}
	
	conf.UploaderCtrlUrl, err = c.String("uploader", "ctrl_url")
	if err != nil && conf.UploaderEnable {
		err = errors.New("Incorrect uploader ctrl_url:" + err.Error())
	}
	
	conf.UploaderParamName, err = c.String("uploader", "token_name")
	if err != nil && conf.UploaderEnable {
		panic("Incorrect uploader token_name:" + err.Error())
	}
	
	
	// Amazon
	conf.AmazonCFEnable, err = c.Bool("amazon", "enable")
	if err != nil {
		conf.AmazonCFEnable = false
	}
	
	if conf.AmazonCFEnable {
		conf.AmazonCFDistributionId, err = c.String("amazon", "distribution_id")
		if err != nil {
			panic(err)
		}
		conf.AmazonCFAuthentication, err = c.String("amazon", "authentication")
		if err != nil {
			panic(err)
		}
		s, err = c.String("amazon", "duration")
		if err != nil {
			conf.AmazonInvalidationDuration = 10 * time.Minute
		} else {
			conf.AmazonInvalidationDuration, err = time.ParseDuration(s)
			if err != nil {
				panic("Incorrect amazon time duration")
			}
			if conf.AmazonInvalidationDuration <= 0 {
				panic("Incorrect amazon time duration")
			} 
		}
	}
	
	err = nil
	
	conf.RegisterMime()
	return
}

// Register all mime types from config
func (conf *Config) RegisterMime() {
	if conf.MimeTypes != nil && len(conf.MimeTypes) > 0 {
		for ext, extType := range conf.MimeTypes {
			mime.AddExtensionType(ext, extType)
		}
	}
}