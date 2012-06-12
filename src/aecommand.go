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

Server addr format:
	default addr: 127.0.0.1:32032
	examples:
		-s=192.168.1.2 will be 192.168.1.2:32032
		-s=:32033 will be 127.0.0.1:32033
		-s=anteater.local:3234 will be anteater.local:3234
		
Commands:
	
	STATUS (or INFO)
	Show remote server info
	aecommand info - show short info
	aecommand info all - show all info
	
	VERSION
	Show remote server version
	aecommand version
	
	DUMP
	Create dump on remote server
	aecommand dump /path/to/dump/folder
	Folder must be exists on remote server
	Server return TRUE or error
	Dump may take some time (depending on your storage size)
`;

const (
	DEFAULT_HOST = "127.0.0.1"
	DEFAULT_PORT = "32032"
)

var (
	ServerAddr *string = flag.String("s", DEFAULT_HOST + ":" + DEFAULT_PORT, "Path to your config file")
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
	
	// check server addr
	addr := strings.Split(*ServerAddr, ":")
	var host, port string
	if len(addr) == 1 {
		host = addr[0]
		port = DEFAULT_PORT
	} else if len(addr) == 2 {
		host = addr[0]
		port = addr[1]
	}
	if len(host) == 0 {
		host = DEFAULT_HOST 
	}
	if len(port) == 0 {
		port = DEFAULT_PORT
	}
	*ServerAddr = host + ":" + port
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
