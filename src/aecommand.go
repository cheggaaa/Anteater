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
	"strings"
	"./anteater"
	"./aerpc"
)

const USAGE = `
Usage:
	aecommand [-s=server_addr] command [arguments]
` + aerpc.SERVER_FLAG_FORMAT + `		
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
	
	CHECKMD5
	Check all files in storage for md5 file hash
`;



var (
	Conn *aerpc.Conn
	Command string
)

func init() {
	var err error
	Conn, err = aerpc.ParseFAndConnect()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	if Conn.ShowHelp {
		fmt.Print(USAGE)
		return
	}
	defer Conn.Client.Close()
	Command = strings.ToLower(Conn.Command)
	Command = strings.ToUpper(string(Command[0])) + Command[1:]
	
	Execute()
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
			err = Conn.Client.Call(command, args, &replyStatus)
			if err != nil {
				log.Fatal("Error:", err)
			}
			
			present := "short"			
			if len(Conn.Args) > 0 {
				present = Conn.Args[0]
			}
			
			res, err := replyStatus.Info(present)
			if err != nil {
				log.Fatal("Error:", err)
			}
			fmt.Printf("%s\n", res)
		case "Ping", "Version", "Checkmd5":
			args := true
			err = Conn.Client.Call(command, args, &replyString)
			if err != nil {
				log.Fatal("Rpc error:", err)
			}
			fmt.Println(replyString)
		case "Dump":
			if len(Conn.Args) == 0 {
				log.Fatalln("Need specified path\nExample: aecommand -s=127.0.0.1:32032 dump /path/to/you/dump/folder")
			}
			args := Conn.Args[0]
			err = Conn.Client.Call(command, args, &replyBool)
			if err != nil {
				log.Fatal("Rpc error:", err)
			}
			fmt.Println(replyBool)
		default:
			log.Fatalf("Undefined command: %s\n", Command)
	}
	return
}
