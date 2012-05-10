package main

import (
	"fire"
	"runtime"
)


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fire.MainInit()
	fire.Start()
	
}
