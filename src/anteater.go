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
)



func main() {
	defer func() {
        if r := recover(); r != nil {
        	fmt.Println("ERROR:", r)
        }
    }()

	fmt.Println(cnst.SIGN)
	c := &config.Config{}
	c.ReadFile("etc/anteater.conf")	
	s := storage.GetStorage(c)
	defer s.Close()
	http.RunServer(s)
		
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGKILL, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	s.Dump()
	fmt.Printf("stopped (%v)\n", sig)
}