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
	"time"
	"runtime"
	"log"
	"runtime/pprof"
)

var config = flag.String("f", "", "Path to your config file")
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

	if *config == "" {
		fmt.Println("Need to specify path to config file\n Use flag -f\n anteater -f /path/to/file.conf")
		return
	}
	
	runtime.GOMAXPROCS(runtime.NumCPU())

	go func() {
		anteater.MainInit(*config)
		anteater.Start()
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Kill, os.Interrupt)
	<-interrupt
	fmt.Println("")
	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.WriteHeapProfile(f)
        f.Close()
        return
    }
	anteater.Log.Debugln("\nCatched shutdown signal...")
	time.Sleep(time.Second)
	anteater.Stop()
}
