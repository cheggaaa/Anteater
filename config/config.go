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
	"errors"
	config "github.com/akrennmair/goconf"
	"github.com/cheggaaa/Anteater/aelog"
	"github.com/cheggaaa/Anteater/utils"
	"mime"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	// Data
	DataPath      string
	ContainerSize int64
	MinEmptySpace int64
	DumpTime      time.Duration
	TmpDir        string
	CpuNum        int

	// Http
	HttpWriteAddr    string
	HttpReadAddr     string
	HttpWriteTimeout time.Duration
	HttpReadTimeout  time.Duration
	ETagSupport      bool
	Md5Header        bool
	ContentRange     int64
	StatusJson       string
	StatusHtml       string

	// Rpc
	RpcAddr string

	// Http Headers
	Headers map[string]string

	// Mime Types
	MimeTypes map[string]string

	// Log
	LogLevel  int
	LogFile   string
	LogAccess string

	//Downloader
	DownloaderEnable    bool
	DownloaderParamName string

	// Uploader
	UploaderEnable    bool
	UploaderCtrlUrl   string
	UploaderParamName string
}

// Parse file and set values to config
func (conf *Config) ReadFile(filename string) {
	c, err := config.ReadConfigFile(filename)
	if err != nil {
		panic(err)
	}

	// Data path
	conf.DataPath, err = c.GetString("data", "path")
	if err != nil {
		panic(err)
	}
	conf.DataPath = strings.TrimRight(conf.DataPath, "/") + "/"

	// Container size
	s, err := c.GetString("data", "container_size")
	if err != nil {
		panic(err)
	}
	conf.ContainerSize, err = utils.BytesFromString(s)

	// Min empty space
	s, err = c.GetString("data", "min_empty_space")
	if err != nil {
		panic(err)
	}
	conf.MinEmptySpace, err = utils.BytesFromString(s)

	// Dump time duration
	s, err = c.GetString("data", "dump_duration")
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
	conf.TmpDir, err = c.GetString("data", "tmp_dir")
	if err == nil {
		conf.TmpDir = strings.TrimRight(conf.TmpDir, "/")
	}

	// Num cpu
	conf.CpuNum, err = c.GetInt("data", "cpu_num")
	if conf.CpuNum < 1 || conf.CpuNum > runtime.NumCPU() {
		conf.CpuNum = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(conf.CpuNum)

	// Http write addr
	conf.HttpWriteAddr, err = c.GetString("http", "write_addr")
	if err != nil {
		panic(err)
	}

	// Http read addr
	conf.HttpReadAddr, err = c.GetString("http", "read_addr")
	if err != nil || len(conf.HttpReadAddr) == 0 {
		conf.HttpReadAddr = conf.HttpWriteAddr
	} else {
		err = nil
	}

	// Http write timeout
	s, err = c.GetString("http", "write_timeout")
	if err == nil && s != "0" && s != "" {
		conf.HttpWriteTimeout, err = time.ParseDuration(s)
		if err != nil {
			panic("Incorrect http.write_timeout time duration")
		}
		if conf.HttpWriteTimeout <= 0 {
			panic("Incorrect http.write_timeout time duration")
		}
	} else {
		err = nil
	}

	// Http read timeout
	s, err = c.GetString("http", "read_timeout")
	if err == nil && s != "0" && s != "" {
		conf.HttpReadTimeout, err = time.ParseDuration(s)
		if err != nil {
			panic("Incorrect http.read_timeout time duration")
		}
		if conf.HttpReadTimeout <= 0 {
			panic("Incorrect http.read_timeout time duration")
		}
	}

	// ETag flag
	conf.ETagSupport, err = c.GetBool("http", "etag")
	if err != nil {
		conf.ETagSupport = true
	}

	// Md5 header
	conf.Md5Header, err = c.GetBool("http", "md5_header")
	if err != nil {
		conf.Md5Header = true
	}

	// Range support
	cr, err := c.GetString("http", "content_range")
	if err != nil {
		cr = "5M"
	}
	conf.ContentRange, err = utils.BytesFromString(cr)

	// Status json
	conf.StatusJson, err = c.GetString("http", "status_json")
	if err != nil {
		conf.StatusJson = ""
	}

	// Status html
	conf.StatusHtml, err = c.GetString("http", "status_html")
	if err != nil {
		conf.StatusHtml = ""
	}

	conf.RpcAddr, err = c.GetString("rpc", "addr")
	if err != nil {
		conf.RpcAddr = ":32032"
	}

	// Headers
	headers := make(map[string]string, 0)
	hOpts, err := c.GetOptions("http-headers")
	if err == nil {
		for _, o := range hOpts {
			v, err := c.GetString("http-headers", o)
			if err == nil && len(v) > 0 {
				headers[o] = v
			}
		}
	}

	conf.Headers = headers

	// Mime
	mimeTypes := make(map[string]string, 0)
	mOpts, err := c.GetOptions("mime-types")
	if err == nil {
		for _, o := range mOpts {
			v, err := c.GetString("mime-types", o)
			if err == nil && len(v) > 0 {
				mimeTypes["."+o] = v
			}
		}
	}

	conf.MimeTypes = mimeTypes

	// Log level
	levels := map[string]int{
		"debug": aelog.LOG_DEBUG,
		"info":  aelog.LOG_INFO,
		"warn":  aelog.LOG_WARN,
	}
	llv, err := c.GetString("log", "level")
	if err != nil {
		llv = "info"
	}
	logLevel, ok := levels[llv]
	if !ok {
		logLevel = levels["info"]
	}
	conf.LogLevel = logLevel

	// Log file
	conf.LogFile, err = c.GetString("log", "file")
	if err != nil {
		conf.LogFile = ""
	}

	// Access log file
	conf.LogAccess, err = c.GetString("log", "access_log")
	if err != nil {
		conf.LogAccess = ""
	}

	// Downloader
	conf.DownloaderEnable, err = c.GetBool("downloader", "enable")
	if err != nil {
		conf.DownloaderEnable = false
	}
	if conf.DownloaderEnable {
		conf.DownloaderParamName, err = c.GetString("downloader", "param_name")
		if err != nil {
			panic("Downloader param_name has error: " + err.Error())
		}
	}

	// Uploader
	conf.UploaderEnable, err = c.GetBool("uploader", "enable")
	if err != nil {
		conf.UploaderEnable = false
	}

	conf.UploaderCtrlUrl, err = c.GetString("uploader", "ctrl_url")
	if err != nil && conf.UploaderEnable {
		err = errors.New("Incorrect uploader ctrl_url:" + err.Error())
	}

	conf.UploaderParamName, err = c.GetString("uploader", "token_name")
	if err != nil && conf.UploaderEnable {
		panic("Incorrect uploader token_name:" + err.Error())
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
