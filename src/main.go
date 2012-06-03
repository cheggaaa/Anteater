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
	"./anteater"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"runtime"
	"log"
	"runtime/pprof"
)

const HELP = anteater.SERVER_SIGN + `
Usage:
	
	-f=/path/to/config/file
	
	-v - print version end exit
	
	-h - show this page
`;

var config = flag.String("f", "", "Path to your config file")
var isPrintVersion = flag.Bool("v", false, "Print version")
var isPrintHelp = flag.Bool("h", false, "Show help")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func init() {
	flag.Parse()
}

func main() {
	if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

	if *isPrintVersion {
		printVersion()
		return
	}
	
	if *isPrintHelp {
		printHelp()
		return
	}

	if *config == "" {
		printHelp()
		return
	}
	
	runtime.GOMAXPROCS(runtime.NumCPU())

	go func() {
		anteater.MainInit(*config)
		anteater.Start()
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGKILL, os.Interrupt, syscall.SIGTERM)
	s := <-interrupt

	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.WriteHeapProfile(f)
        f.Close()
        return
    }
	anteater.Log.Debugln("\nCatched signal", s)
	time.Sleep(time.Second)
	anteater.Stop()
}

func printVersion() {
	fmt.Println(anteater.SERVER_SIGN);
}

func printHelp() {
	fmt.Println(HELP);
}
