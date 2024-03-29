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

package rpcserver

import (
	"github.com/cheggaaa/Anteater/aelog"
	"github.com/cheggaaa/Anteater/backup"
	"github.com/cheggaaa/Anteater/cnst"
	"github.com/cheggaaa/Anteater/stats"
	"github.com/cheggaaa/Anteater/storage"
	"net"
	"net/http"
	"net/rpc"
)

type Storage struct {
	s *storage.Storage
}

func StartRpcServer(s *storage.Storage) {
	r := &Storage{
		s: s,
	}
	rpc.Register(r)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", s.Conf.RpcAddr)
	if e != nil {
		panic("Rpc listen error:" + e.Error())
	}
	aelog.Debugf("Start rpc server on %s\n", s.Conf.RpcAddr)
	go http.Serve(l, nil)
}

func (r *Storage) Status(args *bool, reply *stats.StatsInfo) error {
	*reply = *r.s.GetStats().Info()
	return nil
}

func (r *Storage) Version(args *bool, reply *string) error {
	*reply = cnst.SIGN
	return nil
}

func (r *Storage) Ping(args *bool, reply *string) error {
	*reply = "PONG"
	return nil
}

func (r *Storage) Check(args *bool, reply *error) error {
	*reply = r.s.Check()
	return nil
}

func (r *Storage) Backup(args *string, reply *bool) error {
	*reply = true
	return backup.CreateBackup(r.s, *args)
}

func (r *Storage) FileList(prefix *string, reply *[]string) (err error) {
	*reply, err = r.s.Index.List(*prefix, 0)
	return
}
