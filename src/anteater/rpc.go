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

package anteater

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Rpc int

func StartRpcServer() {
	r := new(Rpc)
	rpc.Register(r)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", Conf.RpcAddr)
	if e != nil {
		log.Fatal("Rpc listen error:", e)
	}
	Log.Debugf("Start rpc server on %s\n", Conf.RpcAddr)
	go http.Serve(l, nil)
}

func (r *Rpc) Status(args *bool, reply **State) error {
	*reply = GetState()
	return nil
}

func (r *Rpc) Version(args bool, reply *string) error {
	*reply = VERSION
	return nil
}

func (r *Rpc) Ping(args bool, reply *string) error {
	*reply = "PONG"
	return nil
}

func (r *Rpc) Reload(args bool, reply *bool) error {
	*reply = true
	return nil
}

func (r *Rpc) Dump(args string, reply *bool) error {
	err := DumpAllTo(args)
	if err != nil {
		*reply = false
		return err
	}
	*reply = true
	return nil
}
