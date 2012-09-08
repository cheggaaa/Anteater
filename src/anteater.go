package main

import (
	"fmt"
	"config"
	"cnst"
)



func main() {
	c := &config.Config{}
	c.ReadFile("etc/anteater.example.conf")
	fmt.Println(cnst.SIGN, c)
}