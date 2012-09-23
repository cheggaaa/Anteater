package main

import (
	"fmt"
	"config"
	"cnst"
	"storage"
	"http"
	"os"
	"os/signal"
	"syscall"
	"aelog"
	"flag"
)

const HELP = cnst.SIGN + `
Usage:
	
	-f=/path/to/config/file
	
	-v - print version end exit
	
	-h - show this page
`;

var configFile = flag.String("f", "", "Path to your config file")
var isPrintVersion = flag.Bool("v", false, "Print version")
var isPrintHelp = flag.Bool("h", false, "Show help")

func init() {
	flag.Parse()
}


func main() {
	defer func() {
        if r := recover(); r != nil {
        	fmt.Println("Error!")
        	fmt.Println(r)
        }
    }()
    
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
	s := storage.GetStorage(c)
	defer s.Close()
	
	
	// init access log is needed
	var al *aelog.AntLog
	if c.LogAccess != "" {
		al, err = aelog.New(c.LogAccess, aelog.LOG_PRINT)
		if err != nil {
			panic(err)
		}
	}
	
	// Run server
	http.RunServer(s, al)
		
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGKILL, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	aelog.Infof("Catched signal %v. Stop server", sig)
	s.Dump()
	aelog.Infoln("")
}


func printVersion() {
	fmt.Println(cnst.SIGN)
}

func printHelp() {
	fmt.Println(HELP)
}