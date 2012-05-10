package anteater

import (
	"net/http"
	"fmt"
	"log"
	"time"
	"sync"
)

var IndexPath string = "/tmp/file.index"
var DataPath  string = "/tmp/file.data"
var CSize int64 = int64(2 * 1024 * 1024 * 1024)
var FreeSpace = int64(300 * 1024 * 1024)

var CleanMutex *sync.Mutex = &sync.Mutex{}


func Init() {
	err := LoadData(IndexPath)
	if err != nil {
		fmt.Println("Error!", err)
		c, err := NewContainer(DataPath, CSize)
		if err != nil {
			fmt.Println("Can't create new container")
			log.Fatal(err)
		}
		FileContainers[c.Id] = c
		Cleanup()
	}
	
	go func() { 
		ch := time.Tick(60 * time.Second)
		for _ = range ch {
			func () {
				defer func() {
			        if err := recover(); err != nil {
			            fmt.Println("Dump failed:", err)
			        }
			    }()
				Cleanup()
			}()
		}
	}()
	
	Start()
}


func Start() {
	runServer(http.HandlerFunc(HttpReadWrite), ":8081")
}


func Stop() {

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
	
	if maxSpace <= FreeSpace {
		c, err := NewContainer(DataPath, CSize)
		if err != nil {
			fmt.Println(err)
		} else {
			FileContainers[c.Id] = c
		}
	}
	
	err := DumpData(IndexPath)
	if err != nil {
		fmt.Println("Dump error:", err)
	}
	
}
