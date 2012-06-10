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
	"log"
	//"time"
	"net/rpc"
	"os"
	"flag"
	"strings"
	"./anteater"
)

const USAGE = `
Usage:
	aecommand [-s=server_addr] command [arguments]
`;

var (
	ServerAddr *string = flag.String("s", ":32032", "Path to your config file")
	Client *rpc.Client
	Command string 
	Args []string
	ShowHelp bool
)

func init() {
	flag.Parse();
	var s int = 1
	if flag.NFlag() == 1 {
		s++
	}
	
	if len(os.Args) - s >= 1 {
		Command = os.Args[s]
		s++
	} else {
		ShowHelp = true
		return
	}
	
	if len(os.Args) - s >= 1 {
		Args = os.Args[s:]
	}
	Command = strings.ToLower(Command)
	Command = strings.ToUpper(string(Command[0])) + Command[1:]
}

func main() {
	if ShowHelp {
		fmt.Print(USAGE)
		return
	}
	//fmt.Printf("Server:%s; Command:%s; args:%v;\n", *ServerAddr, Command, Args);	
	connect();
	defer Client.Close()
	
	Execute()
}

func connect() {
	var err error
	Client, err = rpc.DialHTTP("tcp", *ServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	
}

func Execute() () {
	var command string = "Rpc." + Command
	var err error
	var replyString string
	var replyBool   bool
	var replyStatus *anteater.State
	switch Command {
		case "Status", "Info":
			command = "Rpc.Status"
			args := true
			replyStatus = new(anteater.State)
			err = Client.Call(command, args, &replyStatus)
			if err != nil {
				log.Fatal("Error:", err)
			}
			
			present := "short"			
			if len(Args) > 0 {
				present = Args[0]
			}
			
			res, err := replyStatus.Info(present)
			if err != nil {
				log.Fatal("Error:", err)
			}
			fmt.Printf("%s\n", res)
		case "Ping", "Version":
			args := true
			err = Client.Call(command, args, &replyString)
			if err != nil {
				log.Fatal("Rpc error:", err)
			}
			fmt.Println(replyString)
		case "Dump":
			if len(Args) == 0 {
				log.Fatalln("Need specified path\nExample: aecommand -s=127.0.0.1:32032 dump /path/to/you/dump/folder")
			}
			args := Args[0]
			err = Client.Call(command, args, &replyBool)
			if err != nil {
				log.Fatal("Rpc error:", err)
			}
			fmt.Println(replyBool)
		default:
			log.Fatalf("Undefined command: %s\n", Command)
	}
	return
}
