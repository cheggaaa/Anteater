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

var config string
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func init() {
	flag.StringVar(&config, "f", "", "Path to your config file")
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

	if config == "" {
		fmt.Println("Need to specify path to config file\n Use flag -f\n anteater -f /path/to/file.conf")
		return
	}
	
	runtime.GOMAXPROCS(runtime.NumCPU())

	go func() {
		anteater.Init(config)
		anteater.Start()
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Kill, os.Interrupt)
	<-interrupt
	fmt.Println("")
	anteater.Log.Debugln("\nCatched shutdown signal...")
	time.Sleep(time.Second)
	anteater.Stop()
}
