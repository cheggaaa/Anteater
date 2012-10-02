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
	"testing"
	"os"
	"fmt"
	"aelog"
	"mime"
)




func TestReadFile(t *testing.T) {
	f, err := os.Create("test.conf")
	if err != nil {
		t.Errorf("os.Create has err: %v", err)
	}
	defer os.Remove("test.conf")
	f.Write([]byte(TEST_CONFIG))
	
	c := &Config{}
	c.ReadFile("test.conf")
	
	if configToString(c) != configToString(TestConfig) {
		t.Errorf("TestConfig and Config from file mismatched \nACTUAL:\n%s\nEXPECTED:\n%s\n", configToString(TestConfig), configToString(c))
	}
	
	compare := func (hash, hash2 map[string]string) {
		if len(hash) != len(hash2) {
			t.Errorf("Len mismatched: %d vs %d",  len(hash), len(hash2))
		}
	
		for k, v := range hash {
			if hash2[k] != v {
				t.Errorf("Mismatch: %s (%s vs %s)", k, v, hash2[k])
			}
		}
	}
	
	compare(c.MimeTypes, TestConfig.MimeTypes)
	compare(c.Headers, TestConfig.Headers)
	
	// Check mime register
	if mime.TypeByExtension(".test") != "application/test" {
		t.Error("Mime not registered")
	}
}

var TestConfig *Config = &Config{
	// Data path
	DataPath      : "/opt/DB/anteater/",
	ContainerSize : 200 * 1024,
	MinEmptySpace : 50 * 1024,
	TmpDir        : "/tmp/dir",
	
	HttpWriteAddr : ":8081",
	HttpReadAddr  : ":8080",
	ETagSupport   : true,
	ContentRange  : 5 * 1024,
	StatusJson    : "status.json",
	StatusHtml    : "status.html",
	RpcAddr       : ":32000",
	Headers       : map[string]string{
		"Cache-Control" : "public, max-age=315360000",
	},
	MimeTypes    : map[string]string{
		".jpg"   : "image/jpeg",
		".test"  : "application/test",
	},
	LogLevel     : aelog.LOG_INFO,
	LogFile		 : "/var/log/anteater.log",
	UploaderEnable : true,
	UploaderCtrlUrl : "http://localhost/upload/",
	UploaderParamName : "_token",
	
	DownloaderEnable : true,
	DownloaderParamName : "url",
}

func configToString(c *Config) string {
	return fmt.Sprintln(c.DataPath, c.ContainerSize, c.MinEmptySpace, c.HttpWriteAddr, 
	c.HttpReadAddr, c.ETagSupport, c.Md5Header, c.ContentRange, c.StatusJson, c.StatusHtml, c.RpcAddr, 
	c.LogLevel, c.LogFile, c.UploaderEnable, c.UploaderCtrlUrl, c.UploaderParamName, c.DownloaderEnable, c.DownloaderParamName, c.TmpDir)
}


const TEST_CONFIG = `
[data]

# Path to folder for store files
#path : /path/to/data/folder
path : /opt/DB/anteater

# Size for one file container
container_size : 200K

# Min empty space. After this value will be created new container
min_empty_space : 50K

tmp_dir : /tmp/dir/

[http]

# Addr for listen read requests 
read_addr : :8080

# Addr for listen read and write requests
write_addr : :8081

# ETag support
etag : on

md5_header : off

# Content-Range enable for file biggest then 
# By default it's 5M
content_range : 5K


# Url's for a status page
status_json : status.json
status_html : status.html

[rpc]
addr : :32000

# List of additional http headers
[http-headers]
Cache-Control : public, max-age=315360000

# List of custom mime types, by default use native unix list
[mime-types]
jpg  : image/jpeg 
test : application/test

[log]
# Log level. Should be debug, info or warn
level : info

# File to write log, by default it's stdOut
file  : /var/log/anteater.log

[downloader]
enable : on
param_name : url


[uploader]
# enable or disable uploader
enable     : on

# json api ctrl url
ctrl_url   : http://localhost/upload/

# GET parameter name for uploader
token_name : _token
	
`