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
	"fmt"
	"time"
	"sync"
	"log"
	"mime"
	"crypto/md5"
	"io"
)

const (
	VERSION   = "0.06.5 Narr8"
	SERVER_SIGN = "Anteater " + VERSION
)

/**
 * Path to index file
 **/
var IndexPath string = "file.index"

/**
 * Path to data files
 **/
var DataPath  string = "file.data"


/**
 * Config object
 */
var Conf *Config

/**
 * For Container.Id creation
 */
var ContainerLastId int32

/**
 * Map with container objects
 */
var FileContainers map[int32]*Container = make(map[int32]*Container)

var Containers map[int32]*Container = make(map[int32]*Container)


/**
 *	Mutex for allocate new files
 */
var GetFileLock *sync.Mutex = &sync.Mutex{}

/**
 * File info index
 */
var Index map[string]*FileInfo

/**
 * Lock for Index
 */
var IndexLock *sync.Mutex = &sync.Mutex{}

/**
 * Logger object
 */
var Log *AntLog

/**
 * Server start time
 */
var StartTime time.Time = time.Now()

/**
 * Time of last dump
 */
var LastDump time.Time = time.Now()

/**
 * Making dump time
 */
var LastDumpTime time.Duration

/**
 * Size of index file
 */
var IndexFileSize int64

/**
 * Metrics
 */
var HttpCn  *StateHttpCounters           = &StateHttpCounters{}
var AllocCn *StateAllocateCounters       = &StateAllocateCounters{}


func MainInit(config string) {
	// Init config
	var err error
	Conf, err = LoadConfig(config) 
	if err != nil {
		log.Fatal(err)
	}
	
	// Init logger
	Log, err = LogInit()
	if err != nil {
		log.Fatal(err)
	}
	
	// Set paths
	IndexPath = Conf.DataPath + "/" + IndexPath
	DataPath = Conf.DataPath + "/" + DataPath
	
	
	// Load data from index
	err = LoadData(IndexPath)
	if err != nil {
		// or create new
		Log.Debugln("Error while reading index file:", err)
		Log.Debugln("Try create conainer")
		_, err := NewContainer(DataPath)
		if err != nil {
			Log.Warnln("Can't create new container")
			Log.Fatal(err)
		}
		Cleanup()
	}
	
	go func() { 
		ch := time.Tick(60 * time.Second)
		for _ = range ch {
			func () {
				Cleanup()
			}()
		}
	}()
	
	RegisterMime()
	
	Log.Infoln("Start server with config", config)
}


func Start() {
	StartRpcServer()
	if Conf.HttpReadAddr != Conf.HttpWriteAddr {
		go RunServer(http.HandlerFunc(HttpRead), Conf.HttpReadAddr)
	}
	RunServer(http.HandlerFunc(HttpReadWrite), Conf.HttpWriteAddr)
}


func Stop() {
	Log.Infoln("Server stopping..")
	fmt.Println("Server stopping now")
	Cleanup()
	for _, c := range(FileContainers) {
		c.Disable()
		c.F.Close()
	}
	fmt.Println("Bye")
}

func Cleanup() {
	var maxSpace int64
	var hasChanges bool
	
	for _, c := range(FileContainers) {
		if c.HasChanges() {
			hasChanges = true
		}
		c.Clean()
		if c.MaxSpace() > maxSpace {
			maxSpace = c.MaxSpace()
		}
	}
	
	if maxSpace <= Conf.MinEmptySpace {
		_, err := NewContainer(DataPath)
		if err != nil {
			Log.Warnln(err)
		}
	}
	
	if hasChanges {
		err := DumpData(IndexPath)
		if err != nil {
			Log.Infoln("Dump error:", err)
		}
	}
}


func CheckStorage () (ok bool, description string) {
	ok = true
	for name, fi := range Index {
		Log.Debugf("Check %s ..", name)
		h := md5.New()
		io.Copy(h, fi.GetReader())
		actual := fmt.Sprintf("%x", h.Sum(nil))
		expect := fmt.Sprintf("%x", fi.Md5)
		if actual != expect {
			msg := fmt.Sprintf("File %s md5 not equals! Actual: %s; Expect: %s", name, actual, expect)
			Log.Info(msg)
			description += "\n"
			ok = false
		} else {
			Log.Debugf("Ok")
		}
	}
	if ok {
		description = "Ok!"
	}
	return
}


/**
 * Register Mime types from config
 */
func RegisterMime() {
	types := Conf.MimeTypes
	for k, v := range types {
		mime.AddExtensionType(k, v)
	}
}
