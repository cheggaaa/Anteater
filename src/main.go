package main

import (
	"./anteater"
	"flag"
	"fmt"
)

var config string

func init() {
	flag.StringVar(&config, "f", "", "Path to your config file")
	flag.Parse()
}

func main() {
	fmt.Println(config)
	anteater.Init(config)
	fmt.Println(anteater.Conf)
	anteater.Start()
	anteater.Stop()
}

