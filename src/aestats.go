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
	"fmt"
	"flag"
	"aerpc"
	"aerpc/rpcclient"
	"time"
	"utils"
)

const USAGE = `
Usage:
	aestats [-s="server_addr"]
` + aerpc.SERVER_FLAG_FORMAT + `
`;

var isPrintHelp = flag.Bool("h", false, "Show help")
var serverAddr = flag.String("s", "", "Server addr")

func main() {
	flag.Parse()
	
	aerpc.RegisterCommands()
	
	if *isPrintHelp {
		printHelp()
		return
	}
	
	addr := aerpc.NormalizeAddr(*serverAddr)
	client, err := rpcclient.NewClient(addr)
	if err != nil {
		fmt.Printf("Can't connect to %s:\n%v\n", addr, err)
		return
	}
	defer client.Close()
	var i int
	for {
		cmd := new(aerpc.RpcCommandStatus)
		err = cmd.Execute(client)
		if err != nil {
			panic(err)
			return
		}
		if i % 10 == 0 {
			printHead()
		}
		stat := cmd.Data().(*aerpc.RpcCommandStatus)
		printStats(stat)
		time.Sleep(time.Second)
		i++
	}
}

func printHead() {
	fmt.Printf("Get\tAdd\tDel\tNF\tOP\tHP\tHPS\tIn / Out\n")
}

var old *aerpc.RpcCommandStatus

func printStats(stat *aerpc.RpcCommandStatus) {
	if old != nil {
		get := stat.Counters["get"] - old.Counters["get"]
		add := stat.Counters["add"] - old.Counters["add"]
		del := stat.Counters["delete"] - old.Counters["delete"]
		notfound := stat.Counters["notFound"] - old.Counters["notFound"]
		op  := stat.Storage.IndexVersion - old.Storage.IndexVersion
		
		hp  := float64(stat.Storage.HoleCount) / float64(stat.Storage.FilesCount) * 100 
		hps := float64(stat.Storage.HoleSize) / float64(stat.Storage.FilesSize) * 100 
		
		in := stat.Traffic["in"] - old.Traffic["in"]
		out := stat.Traffic["out"] - old.Traffic["out"]
		
		fmt.Printf("%d\t%d\t%d\t%d\t%d\t%.2f%%\t%.2f%%\t%s / %s\n", get, add, del, notfound, op, hp, hps, utils.HumanBytes(int64(in)), utils.HumanBytes(int64(out)))
	}
	old = stat
}


func printHelp() {
	fmt.Print(USAGE)
}