package anteater

import (
	"net/http"
	"fmt"
	"time"
	"sync"
	"log"
)

const (
	version   = "0.02.1"
	serverSign = "AE " + version
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
	DataPath = Conf.DataPath + "/ " + DataPath
	
	
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
	for _, c := range(FileContainers) {
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
