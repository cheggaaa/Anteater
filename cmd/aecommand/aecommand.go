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
	"flag"
	"fmt"
	"github.com/cheggaaa/Anteater/aerpc"
	"github.com/cheggaaa/Anteater/aerpc/rpcclient"
	"os"
	"strings"
)

const USAGE = `
Usage:
	aecommand [-s="server_addr"] command [arguments]
` + aerpc.SERVER_FLAG_FORMAT + `		
Commands:
	%s
`

var isPrintHelp = flag.Bool("h", false, "Show help")
var serverAddr = flag.String("s", "", "Server addr")

func main() {
	flag.Parse()

	aerpc.RegisterCommands()

	if *isPrintHelp {
		printHelp()
		return
	}

	command, args := parseArgs()

	if command == "" {
		fmt.Println("Command not specified")
		return
	}

	cmd, ok := aerpc.Commands[command]
	if !ok {
		fmt.Printf("Call to undefined command: %s\n", command)
		return
	}

	addr := aerpc.NormalizeAddr(*serverAddr)
	client, err := rpcclient.NewClient(addr)
	if err != nil {
		fmt.Printf("Can't connect to %s:\n%v\n", addr, err)
		return
	}
	defer client.Close()

	if args != nil && len(args) > 0 {
		err = cmd.SetArgs(args)
		if err != nil {
			fmt.Printf("Invalid args format: %v\n", err)
			return
		}
	}
	err = cmd.Execute(client)
	if err != nil {
		fmt.Printf("Server return error:\n%v\n", err)
	} else {
		cmd.Print()
	}
}

func parseArgs() (command string, args []string) {
	s := flag.NFlag() + 1

	if len(os.Args)-s >= 1 {
		command = os.Args[s]
		s++
	}

	if len(os.Args)-s >= 1 {
		args = os.Args[s:]
	}

	command = strings.ToUpper(command)

	return
}

func printHelp() {
	var commandsHelp = ""

	for name, cmd := range aerpc.Commands {
		commandsHelp += fmt.Sprintf("\n%s\n%s\n", name, cmd.Help())
	}
	fmt.Printf(USAGE, commandsHelp)
}
