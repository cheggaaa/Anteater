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

package main

import (
	"flag"
	"fmt"
	"github.com/cheggaaa/Anteater/aelog"
	"github.com/cheggaaa/Anteater/aerpc/rpcserver"
	"github.com/cheggaaa/Anteater/cnst"
	"github.com/cheggaaa/Anteater/config"
	"github.com/cheggaaa/Anteater/http"
	"github.com/cheggaaa/Anteater/storage"
	"log"
	ghttp "net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

const HELP = cnst.SIGN + `
Usage:
	
	-f=/path/to/config/file
	
	-v - print version end exit
	
	-h - show this page
`

var configFile = flag.String("f", "", "Path to your config file")
var isPrintVersion = flag.Bool("v", false, "Print version")
var isPrintHelp = flag.Bool("h", false, "Show help")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var httpprofile = flag.Bool("httpprofile", false, "enable http profiling")

func init() {
	flag.Parse()
}

func main() {
	defer func() {
		/* if r := recover(); r != nil {
		   	fmt.Println("Error!")
		   	fmt.Println(r)
		   }
		*/
	}()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *httpprofile {
		go func() {
			log.Println(ghttp.ListenAndServe(":6060", nil))
		}()
	}

	var err error

	if *isPrintVersion {
		printVersion()
		return
	}

	if *isPrintHelp {
		printHelp()
		return
	}

	// If configFile not specified - show help
	if *configFile == "" {
		printHelp()
		return
	}

	// Init config
	c := &config.Config{}
	c.ReadFile(*configFile)

	// Init logger
	aelog.DefaultLogger, err = aelog.New(c.LogFile, c.LogLevel)
	if err != nil {
		panic(err)
	}

	// Init storage
	stor := &storage.Storage{}
	stor.Init(c)
	err = stor.Open()
	if err != nil {
		panic(err)
	}
	defer stor.Close()

	// init access log is needed
	var al *aelog.AntLog
	if c.LogAccess != "" {
		al, err = aelog.New(c.LogAccess, aelog.LOG_PRINT)
		if err != nil {
			panic(err)
		}
	}

	// Run server
	http.RunServer(stor, al)

	// Run rpc server
	rpcserver.StartRpcServer(stor)

	aelog.Infof("Run working (use %d cpus)", c.CpuNum)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGKILL, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	aelog.Infof("Catched signal %v. Stop server", sig)
	stor.Dump()
	aelog.Infoln("")
}

func printVersion() {
	fmt.Println(cnst.SIGN)
}

func printHelp() {
	fmt.Println(HELP)
}
