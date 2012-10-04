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

package aerpc

import (
	"fmt"
	"net/rpc"
	"stats"
	"utils"
)

// Interface
type RpcCommand interface {
	ShortName() string
	RpcName() string
	Help() string
	SetArgs(args []string)
	Execute(client *rpc.Client) error
	Print()
}

// All rpc commands as map[cmd_name] cmd
var Commands  = make(map[string]RpcCommand)

func RegisterCommands() {
	var cmds = make([]RpcCommand, 0)	
	cmds = append(cmds, new(RpcCommandVersion))
	cmds = append(cmds, new(RpcCommandPing))
	cmds = append(cmds, new(RpcCommandStatus))
	cmds = append(cmds, new(RpcCommandCheckMD5))
	
	for _, cmd := range cmds {
		Commands[cmd.ShortName()] = cmd
	}
}

// VERSION
type RpcCommandVersion string
func (c *RpcCommandVersion) ShortName() string { return "VERSION" }
func (c *RpcCommandVersion) RpcName() string { return "Storage.Version" }
func (c *RpcCommandVersion) Help() string { return "Print Anteater server version" }
func (c *RpcCommandVersion) SetArgs(args []string) { return }
func (c *RpcCommandVersion) Print() { fmt.Println(*c) }
func (c *RpcCommandVersion) Execute(client *rpc.Client) (err error) {
	err = client.Call(c.RpcName(), true, c)
	return
}

// PING
type RpcCommandPing string
func (c *RpcCommandPing) ShortName() string { return "PING" }
func (c *RpcCommandPing) RpcName() string { return "Storage.Ping" }
func (c *RpcCommandPing) Help() string { return "Return PONG if server running" }
func (c *RpcCommandPing) SetArgs(args []string) { return }
func (c *RpcCommandPing) Print() { fmt.Println(*c) }
func (c *RpcCommandPing) Execute(client *rpc.Client) (err error) {
	err = client.Call(c.RpcName(), true, c)
	return
}

// STATUS
type RpcCommandStatus stats.StatsInfo
func (c *RpcCommandStatus) ShortName() string { return "STATUS" }
func (c *RpcCommandStatus) RpcName() string { return "Storage.Status" }
func (c *RpcCommandStatus) Help() string { return "Return server status info" }
func (c *RpcCommandStatus) SetArgs(args []string) { return }
func (c *RpcCommandStatus) Print() { 
	fmt.Println("Anteater")
	fmt.Printf("  Start time: %v\n  Version: %s\n\n", c.Anteater.StartTime, c.Anteater.Version)
	fmt.Println("Enviroment")
	fmt.Printf("  Go version: %s\n  Server time:  %v\n  Num goroutines: %d\n  Memory allocated: %s\n\n", c.Env.GoVersion, c.Env.Time, c.Env.NumGoroutine, utils.HumanBytes(int64(c.Env.MemAlloc)))
	fmt.Println("Storage")
	fmt.Printf("  Containers count: %d\n  Files count: %d\n  Files size: %s\n  Holes: %s (%d)\n  Index version: %d\n\n", 
		c.Storage.ContainersCount, c.Storage.FilesCount, utils.HumanBytes(c.Storage.FilesSize), 
		utils.HumanBytes(c.Storage.HoleSize), c.Storage.HoleCount, c.Storage.IndexVersion)
	fmt.Println("Counters")
	fmt.Printf("  Get: %d\n  Add: %d\n  Delete: %d\n  Not found: %d\n  Not modified: %d\n\n", 
		c.Counters["get"], c.Counters["add"], c.Counters["delete"], c.Counters["notFound"], c.Counters["notModified"])
	fmt.Println("Traffic")
	fmt.Printf("  In: %s\n  Out: %s\n\n", utils.HumanBytes(int64(c.Traffic["in"])), utils.HumanBytes(int64(c.Traffic["out"])))
	fmt.Println("Allocates")
	fmt.Printf("  Append: %d\n  Replace: %d\n  In hole: %d\n\n", c.Allocate["append"], c.Allocate["replace"], c.Allocate["in"])
}
func (c *RpcCommandStatus) Execute(client *rpc.Client) (err error) {
	err = client.Call(c.RpcName(), true, c)
	return
}

// CHECK MD5
type RpcCommandCheckMD5 map[string]bool
func (c *RpcCommandCheckMD5) ShortName() string { return "CHECKMD5" }
func (c *RpcCommandCheckMD5) RpcName() string { return "Storage.CheckMD5" }
func (c *RpcCommandCheckMD5) Help() string { return "Return PONG if server running" }
func (c *RpcCommandCheckMD5) SetArgs(args []string) { return }
func (c *RpcCommandCheckMD5) Print() {
	var ok, e int
	for n, r := range *c {
		if ! r {
			fmt.Printf("File: %s has md5 error\n", n)
			e++
		} else {
			ok++
		}
	}
	fmt.Printf("Total scaned: %d. Errors: %d\n", len(*c), e)
}
func (c *RpcCommandCheckMD5) Execute(client *rpc.Client) (err error) {
	err = client.Call(c.RpcName(), true, c)
	return
}