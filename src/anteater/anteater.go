package anteater

import (
	"net/http"
	"fmt"
	"log"
	"time"
	"sync"
)

var IndexPath string = "file.index"
var DataPath  string = "file.data"
var CSize int64
var FreeSpace int64

var CleanMutex *sync.Mutex = &sync.Mutex{}


func Init(config string) {
	err := LoadConfig(config) 
	if err != nil {
		log.Fatal(err)
	}
	
	err = LogInit()
	if err != nil {
		log.Fatal(err)
	}
	
	IndexPath = Conf.DataPath + "/" + IndexPath
	DataPath = Conf.DataPath + "/ " + DataPath
	CSize = Conf.ContainerSize
	
	err = LoadData(IndexPath)
	if err != nil {
		Log.Debugln("Error while reading index file:", err)
		Log.Debugln("Try create conainer")
		c, err := NewContainer(DataPath, CSize)
		if err != nil {
			Log.Warnln("Can't create new container")
			Log.Fatal(err)
		}
		FileContainers[c.Id] = c
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
	
	Log.Infoln("Start server with config", config)
}


func Start() {
	if Conf.HttpReadAddr != Conf.HttpWriteAddr {
		go RunServer(http.HandlerFunc(HttpRead), Conf.HttpReadAddr)
	}
	RunServer(http.HandlerFunc(HttpReadWrite), Conf.HttpWriteAddr)
}


func Stop() {
	Log.Infoln("Server stopping..")
	fmt.Println("Server stopping now")
	Cleanup()
	fmt.Println("Bye")
}

func Cleanup() {
	GetFileLock.Lock()
	defer GetFileLock.Unlock()
	var maxSpace int64
	for _, c := range(FileContainers) {
		c.Clean()
		if c.MaxSpace() > maxSpace {
			maxSpace = c.MaxSpace()
		}
	}
	
	if maxSpace <= Conf.MinEmptySpace {
		c, err := NewContainer(DataPath, CSize)
		if err != nil {
			Log.Warnln(err)
		} else {
			FileContainers[c.Id] = c
		}
	}
	
	err := DumpData(IndexPath)
	if err != nil {
		Log.Infoln("Dump error:", err)
	}
	
}
